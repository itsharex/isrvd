#!/bin/bash
set -e

# ------------------------------------------
# isrvd 服务管理脚本
# 用法: isrvd.sh [install|update|uninstall|download]
# ------------------------------------------

# 配置
SERVICE_NAME="isrvd"
INSTALL_DIR="/usr/local/isrvd"
CONFIG_FILE="$INSTALL_DIR/config.yml"
BIN_LINK="/usr/local/bin/isrvd"
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"

# ------------------------------------------
# 版本信息
# ------------------------------------------

get_arch() {
    case "$(uname -m)" in
        x86_64|amd64) echo "amd64" ;;
        aarch64|arm64) echo "arm64" ;;
        armv7l|armhf) echo "arm" ;;
        *) uname -m ;;
    esac
}

get_latest_version() {
    curl -sI https://github.com/rehiy/isrvd/releases/latest 2>/dev/null | \
        grep -i "location:" | sed 's#.*/tag/##' | tr -d '\r\n' || echo "unknown"
}

ARCH=$(uname -s | tr '[:upper:]' '[:lower:]')-$(get_arch)
LATEST=$(get_latest_version)
DOWNLOAD_URL="https://github.com/rehiy/isrvd/releases/download/$LATEST/isrvd-$ARCH.tar.gz"

# ------------------------------------------
# systemctl 服务管理
# ------------------------------------------

service_install() {
    local bin_file=$(get_bin_file)

    cat > "$SERVICE_FILE" << EOF
[Unit]
Description=isrvd - Infrastructure Service Daemon
Documentation=https://github.com/rehiy/isrvd
After=network.target docker.service
Wants=docker.service

[Service]
Type=simple
User=root
WorkingDirectory=$INSTALL_DIR
Environment="CONFIG_PATH=$CONFIG_FILE"
ExecStart=$bin_file
ExecReload=/bin/kill -HUP \$MAINPID
Restart=on-failure
RestartSec=5s

# 安全加固
NoNewPrivileges=true
LimitNOFILE=65536

# 日志
StandardOutput=journal
StandardError=journal
SyslogIdentifier=isrvd

[Install]
WantedBy=multi-user.target
EOF
    echo "[info] Installing systemd service"
    systemctl daemon-reload
    systemctl enable "$SERVICE_NAME"
}

service_uninstall() {
    if [ ! -f "$SERVICE_FILE" ]; then
        return
    fi
    if systemctl is-enabled --quiet "$SERVICE_NAME" 2>/dev/null; then
        echo "[info] Disabling ${SERVICE_NAME} service..."
        systemctl disable "$SERVICE_NAME"
    fi
    echo "[info] Removing service file"
    rm -f "$SERVICE_FILE"
    systemctl daemon-reload
}

service_start() {
    echo "[info] Starting ${SERVICE_NAME} service..."
    systemctl start "$SERVICE_NAME"
}

service_stop() {
    if systemctl is-active --quiet "$SERVICE_NAME"; then
        echo "[info] Stopping ${SERVICE_NAME} service..."
        systemctl stop "$SERVICE_NAME"
    fi
}

# ------------------------------------------
# 文件管理
# ------------------------------------------

get_bin_file() {
    find "$INSTALL_DIR" -type f -name "isrvd-*" -executable | head -1
}

files_install() {
    echo "[info] Creating directory: $INSTALL_DIR"
    mkdir -p "$INSTALL_DIR"

    echo "[info] Extracting to: $INSTALL_DIR"
    tar xzf "$1" -C "$INSTALL_DIR"

    # 创建命令链接
    local bin_file=$(get_bin_file)
    echo "[info] Creating symlink: $BIN_LINK -> $bin_file"
    ln -sf "$bin_file" "$BIN_LINK"
}

files_update() {
    echo "[info] Extracting to: $INSTALL_DIR"
    tar xzf "$1" -C "$INSTALL_DIR"

    # 更新符号链接
    local bin_file=$(get_bin_file)
    if [ ! -L "$BIN_LINK" ] || [ "$(readlink -f "$BIN_LINK")" != "$bin_file" ]; then
        echo "[info] Updating symlink: $BIN_LINK -> $bin_file"
        ln -sf "$bin_file" "$BIN_LINK"
    fi
}

files_uninstall() {
    if [ -L "$BIN_LINK" ]; then
        echo "[info] Removing symlink: $BIN_LINK"
        rm -f "$BIN_LINK"
    fi

    if [ -d "$INSTALL_DIR" ]; then
        echo "[info] Removing directory: $INSTALL_DIR"
        rm -rf "$INSTALL_DIR"
    fi
}

# ------------------------------------------
# 下载
# ------------------------------------------

download_package() {
    local tmp_file

    tmp_file=$(mktemp)
    echo "[info] Downloading: $DOWNLOAD_URL"

    if ! curl -fL --progress-bar -o "$tmp_file" "$DOWNLOAD_URL"; then
        echo "[error] Download failed"
        rm -f "$tmp_file"
        return 1
    fi

    DOWNLOAD_TMP_FILE="$tmp_file"
}

# ------------------------------------------
# 主流程
# ------------------------------------------

check_root() {
    if [ "$EUID" -ne 0 ]; then
        echo "[error] Please run as root"
        exit 1
    fi
}

install() {
    check_root

    echo "=========================================="
    echo "  isrvd Installer"
    echo "=========================================="
    echo "  Version: $LATEST"
    echo "  Arch:    $ARCH"
    echo "=========================================="

    if [ -d "$INSTALL_DIR" ]; then
        echo "[error] Already installed: $INSTALL_DIR"
        echo "[info] Use 'update' to upgrade"
        exit 1
    fi

    if ! download_package; then
        exit 1
    fi

    files_install "$DOWNLOAD_TMP_FILE"
    service_install
    service_start

    rm -f "$DOWNLOAD_TMP_FILE"

    local bin_file=$(get_bin_file)

    echo ""
    echo "=========================================="
    echo "  Installation Complete!"
    echo "=========================================="
    echo "  Install: $INSTALL_DIR"
    echo "  Binary:  $bin_file"
    echo "  Config:  $CONFIG_FILE"
    echo ""
    echo "  Commands:"
    echo "    systemctl start $SERVICE_NAME"
    echo "    systemctl stop $SERVICE_NAME"
    echo "    systemctl status $SERVICE_NAME"
    echo "    journalctl -u $SERVICE_NAME -f"
    echo "=========================================="
}

update() {
    check_root

    echo "=========================================="
    echo "  isrvd Updater"
    echo "=========================================="
    echo "  Latest: $LATEST"
    echo "=========================================="

    if [ ! -d "$INSTALL_DIR" ]; then
        echo "[error] Not installed"
        echo "[info] Use 'install' first"
        exit 1
    fi

    service_stop

    if ! download_package; then
        service_start
        exit 1
    fi

    files_update "$DOWNLOAD_TMP_FILE"
    service_start

    rm -f "$DOWNLOAD_TMP_FILE"

    local bin_file=$(get_bin_file)

    echo ""
    echo "=========================================="
    echo "  Update Complete!"
    echo "=========================================="
    echo "  Install: $INSTALL_DIR"
    echo "  Binary:  $bin_file"
    echo "  Config:  $CONFIG_FILE (preserved)"
    echo "=========================================="
}

uninstall() {
    check_root

    echo "=========================================="
    echo "  isrvd Uninstaller"
    echo "=========================================="

    service_stop
    service_uninstall
    files_uninstall

    echo ""
    echo "=========================================="
    echo "  Uninstallation Complete!"
    echo "=========================================="
}

download() {
    echo "=========================================="
    echo "  isrvd Downloader"
    echo "=========================================="
    echo "  Version: $LATEST"
    echo "  Arch:    $ARCH"
    echo "  URL:     $DOWNLOAD_URL"
    echo "=========================================="

    if ! download_package; then
        exit 1
    fi

    mv "$DOWNLOAD_TMP_FILE" "isrvd-$ARCH.tar.gz"

    echo ""
    echo "[info] Downloaded: isrvd-$ARCH.tar.gz"
    ls -la "isrvd-$ARCH.tar.gz"
}

# ------------------------------------------
# 入口
# ------------------------------------------

case "${1:-}" in
    install)   install ;;
    update)    update ;;
    uninstall) uninstall ;;
    download)  download ;;
    *)
        echo "Usage: $0 {install|update|uninstall|download}"
        echo ""
        echo "Commands:"
        echo "  install   - Download and install isrvd service"
        echo "  update    - Update to latest version"
        echo "  uninstall - Remove isrvd service and config"
        echo "  download  - Download latest package to current directory"
        echo ""
        echo "Latest version: $LATEST"
        exit 1
        ;;
esac