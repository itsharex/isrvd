package server

import (
	"context"
	"io"
	"os/exec"
	"runtime"

	"github.com/creack/pty"
	"github.com/gin-gonic/gin"
	"github.com/rehiy/libgo/command"
	"github.com/rehiy/libgo/logman"
	"github.com/rehiy/libgo/websocket"
)

// defineShellRoutes 定义 Shell 模块路由（Web 终端）
func (app *App) defineShellRoutes() []Route {
	return []Route{
		{Method: "GET", Path: "/shell", Handler: app.shellWebSocket, Module: "shell", Label: "打开 Web Shell 终端"},
	}
}

func (app *App) shellWebSocket(c *gin.Context) {
	username := c.GetString("username")
	member := app.accountSvc.MemberInspect(username)
	if member == nil {
		logman.Error("用户不存在", "username", username)
		c.AbortWithStatus(403)
		return
	}

	shell := c.DefaultQuery("shell", command.DefaultShell())

	// 使用 Handler 模式处理 WebSocket
	app.wsConfig.Handler(func(conn *websocket.ServerConn) {
		shellRunTerminal(conn, shell, member.HomeDirectory)
	})(c)
}

func shellRunTerminal(conn *websocket.ServerConn, shell, homeDir string) {
	shell = command.GetShell(shell)

	ctx := context.Background()

	// PTY 模式（仅非 Windows）
	if runtime.GOOS != "windows" {
		cmd := command.NewCommand(ctx, shell, nil, homeDir)
		ptmx, err := pty.Start(cmd)
		if err == nil {
			defer ptmx.Close()
			shellHandleIO(conn, ptmx, ptmx, cmd)
			return
		}
		logman.Warn("PTY 启动失败，降级到 Pipe 模式", "error", err)
		conn.Write([]byte("[提示: PTY 模式不可用，已降级到 Pipe 模式]\r\n"))
	}

	// Pipe 模式
	cmd := command.NewCommand(ctx, shell, nil, homeDir)
	if err := shellRunWithPipe(conn, cmd); err != nil {
		logman.Error("Pipe 模式启动失败", "shell", shell, "error", err)
		conn.Write([]byte("[启动 " + shell + " 失败: " + err.Error() + "]\r\n"))
	}
}

func shellRunWithPipe(conn *websocket.ServerConn, cmd *exec.Cmd) error {
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	defer stdin.Close()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		return err
	}
	shellHandleIO(conn, stdin, stdout, cmd)
	return nil
}

func shellHandleIO(conn *websocket.ServerConn, stdin io.Writer, stdout io.Reader, cmd *exec.Cmd) {
	// 确保进程被终止
	defer func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
			cmd.Wait()
		}
	}()

	// 读取输出
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				conn.Write(buf[:n])
			}
			if err != nil {
				logman.Error("shellHandleIO: stdout.Read error", "error", err)
				return
			}
		}
	}()

	conn.Write([]byte("[终端已连接，输入命令后回车]\r\n"))

	// 读取输入
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			logman.Error("shellHandleIO: conn.Read error", "error", err)
			return
		}
		if n > 0 {
			if _, err = stdin.Write(buf[:n]); err != nil {
				logman.Error("shellHandleIO: stdin.Write error", "error", err)
				return
			}
		}
	}
}
