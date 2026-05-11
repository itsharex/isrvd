package docker

import (
	"context"
	"io"

	"github.com/docker/docker/api/types/container"
	"github.com/rehiy/libgo/logman"
	"github.com/rehiy/libgo/websocket"
)

// ContainerExec 容器终端 WebSocket 处理（业务逻辑层）
func (s *DockerService) ContainerExec(ctx context.Context, conn *websocket.ServerConn, containerID, shell string) {
	if shell == "" {
		shell = "/bin/sh"
	}

	// 创建 exec 实例
	execConfig := container.ExecOptions{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Cmd:          []string{shell},
	}

	execResp, err := s.client.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		conn.Write([]byte("[创建终端会话失败: " + err.Error() + "]\r\n"))
		return
	}

	// 连接到 exec 实例
	attachConfig := container.ExecStartOptions{Tty: true}
	hijackedResp, err := s.client.ContainerExecAttach(ctx, execResp.ID, attachConfig)
	if err != nil {
		conn.Write([]byte("[连接终端失败: " + err.Error() + "]\r\n"))
		return
	}
	defer hijackedResp.Close()

	conn.Write([]byte("[容器终端已连接]\r\n"))

	// 转发容器输出到 WebSocket
	done := make(chan struct{})
	go func() {
		defer close(done)
		buf := make([]byte, 1024)
		for {
			n, err := hijackedResp.Reader.Read(buf)
			if err != nil {
				if err != io.EOF {
					logman.Error("Container exec read error", "error", err)
				}
				return
			}
			if n > 0 {
				conn.Write(buf[:n])
			}
		}
	}()

	// 转发 WebSocket 输入到容器
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			logman.Error("WebSocket read error", "error", err)
			break
		}
		if n > 0 {
			if _, err := hijackedResp.Conn.Write(buf[:n]); err != nil {
				logman.Error("Container exec write error", "error", err)
				break
			}
		}
	}

	// 关闭 hijack 连接触发 reader goroutine 退出，等待其结束后函数返回
	hijackedResp.Close()
	<-done
}
