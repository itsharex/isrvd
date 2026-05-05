#!/usr/bin/env python3
"""
前端代码批量格式化脚本
功能：安装依赖、格式化代码、整理 import、类型检查
"""

import subprocess
import sys
from pathlib import Path


# 颜色输出
class Color:
    RED = '\033[0;31m'
    GREEN = '\033[0;32m'
    YELLOW = '\033[1;33m'
    BLUE = '\033[0;34m'
    NC = '\033[0m'


def log_info(msg):
    print(f"{Color.BLUE}[INFO]{Color.NC} {msg}")


def log_success(msg):
    print(f"{Color.GREEN}[SUCCESS]{Color.NC} {msg}")


def log_warn(msg):
    print(f"{Color.YELLOW}[WARN]{Color.NC} {msg}")


def log_error(msg):
    print(f"{Color.RED}[ERROR]{Color.NC} {msg}")


def run_command(cmd, check=True, capture=False):
    """运行命令"""
    try:
        if capture:
            result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
            return result.returncode == 0, result.stdout, result.stderr
        else:
            result = subprocess.run(cmd, shell=True)
            return result.returncode == 0, None, None
    except Exception as e:
        if check:
            log_error(f"命令执行失败: {e}")
            sys.exit(1)
        return False, None, str(e)


def check_npm():
    """检查 npm 是否可用"""
    success, _, _ = run_command("npm --version", check=False, capture=True)
    if not success:
        log_error("npm 未安装，请先安装 Node.js")
        sys.exit(1)


def install_prettier():
    """安装 Prettier（如果未安装）"""
    log_info("检查 Prettier 安装状态...")
    
    success, _, _ = run_command("npm list prettier", check=False, capture=True)
    if not success:
        log_info("安装 Prettier 及相关插件...")
        success, _, err = run_command(
            "npm install --save-dev prettier @prettier/plugin-phtml prettier-plugin-tailwindcss",
            check=False
        )
        if not success:
            log_warn("Prettier 插件安装失败，使用基础配置")
            run_command("npm install --save-dev prettier")
        log_success("Prettier 安装完成")
    else:
        log_success("Prettier 已安装")


def create_prettier_config():
    """创建 Prettier 配置文件"""
    root_dir = Path(__file__).parent.parent
    config_file = root_dir / ".prettierrc"
    ignore_file = root_dir / ".prettierignore"
    
    if not config_file.exists():
        log_info("创建 Prettier 配置文件...")
        config_file.write_text("""{
  "semi": true,
  "singleQuote": true,
  "tabWidth": 2,
  "trailingComma": "es5",
  "printWidth": 100,
  "bracketSpacing": true,
  "arrowParens": "always",
  "endOfLine": "lf",
  "vueIndentScriptAndStyle": false,
  "htmlWhitespaceSensitivity": "css",
  "plugins": ["@prettier/plugin-phtml", "prettier-plugin-tailwindcss"]
}
""")
        log_success("Prettier 配置文件已创建")
    else:
        log_info("Prettier 配置文件已存在")
    
    if not ignore_file.exists():
        ignore_file.write_text("""node_modules
dist
*.min.js
*.min.css
package-lock.json
""")
        log_success(".prettierignore 已创建")


def format_code(dry_run=False):
    """格式化代码"""
    log_info("开始格式化代码...")
    
    patterns = ["src/**/*.vue", "src/**/*.ts", "src/**/*.js"]
    
    for pattern in patterns:
        if dry_run:
            cmd = f"npx prettier --check '{pattern}'"
        else:
            cmd = f"npx prettier --write '{pattern}'"
            log_info(f"格式化: {pattern}")
        
        success, _, err = run_command(cmd, check=False, capture=True)
        if not success and not dry_run:
            log_warn(f"部分文件格式化失败: {pattern}")
    
    if not dry_run:
        log_success("代码格式化完成")


def sort_imports(dry_run=False):
    """整理 import 语句"""
    log_info("整理 import 语句...")
    
    script_dir = Path(__file__).parent
    sort_script = script_dir / "sort-imports.py"
    
    if sort_script.exists():
        cmd = f"python3 {sort_script}"
        if dry_run:
            cmd += " --dry-run"
        run_command(cmd)
        log_success("import 语句整理完成")
    else:
        log_warn("sort-imports.py 脚本不存在，跳过")


def type_check():
    """类型检查"""
    log_info("执行 TypeScript 类型检查...")
    
    success, _, _ = run_command("npm run lint", check=False)
    if success:
        log_success("类型检查通过")
    else:
        log_warn("类型检查发现问题，请手动检查")


def show_help():
    """显示帮助"""
    print("""用法: python format.py [选项]

选项:
  --install     仅安装 Prettier
  --format      仅执行格式化
  --imports     仅整理 import
  --check       仅执行类型检查
  --all         执行所有步骤（默认）
  --dry-run     预览模式（不修改文件）
  -h, --help    显示帮助信息

示例:
  python format.py                  # 执行完整格式化流程
  python format.py --format         # 仅格式化代码
  python format.py --dry-run        # 预览将要修改的文件
""")


def main():
    action = "all"
    dry_run = False
    
    # 解析参数
    args = sys.argv[1:]
    for arg in args:
        if arg == "--install":
            action = "install"
        elif arg == "--format":
            action = "format"
        elif arg == "--imports":
            action = "imports"
        elif arg == "--check":
            action = "check"
        elif arg == "--all":
            action = "all"
        elif arg == "--dry-run":
            dry_run = True
        elif arg in ["-h", "--help"]:
            show_help()
            sys.exit(0)
        else:
            log_error(f"未知参数: {arg}")
            show_help()
            sys.exit(1)
    
    log_info("前端代码格式化工具")
    log_info("========================")
    
    check_npm()
    
    if action == "install":
        install_prettier()
        create_prettier_config()
    elif action == "format":
        install_prettier()
        create_prettier_config()
        format_code(dry_run)
    elif action == "imports":
        sort_imports(dry_run)
    elif action == "check":
        type_check()
    elif action == "all":
        install_prettier()
        create_prettier_config()
        if dry_run:
            log_info("预览模式：检查需要格式化的文件...")
            format_code(dry_run=True)
            sort_imports(dry_run=True)
        else:
            format_code()
            sort_imports()
            type_check()
    
    log_success("完成！")


if __name__ == "__main__":
    main()
