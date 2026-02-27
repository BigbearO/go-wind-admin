#!/usr/bin/env bash
set -euo pipefail

####################################
## 自动检测操作系统并调用对应的准备脚本
####################################

# 获取当前脚本所在目录的绝对路径
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

log() { echo "==> $*"; }
err_trap() { echo "错误：第 $1 行发生错误" >&2; exit 1; }
trap 'err_trap $LINENO' ERR

# 检测操作系统
detect_os() {
  local os_type=""

  # 检测 macOS
  if [[ "$OSTYPE" == "darwin"* ]]; then
    os_type="macos"
  # 检测 Linux 发行版
  elif [[ "$OSTYPE" == "linux-gnu"* ]] || [[ "$OSTYPE" == "linux"* ]]; then
    if [ -f /etc/os-release ]; then
      . /etc/os-release
      case "$ID" in
        ubuntu|debian|linuxmint|pop)
          os_type="ubuntu"
          ;;
        centos|rhel)
          if [ "${VERSION_ID%%.*}" -ge 8 ] 2>/dev/null; then
            os_type="centos"
          else
            os_type="centos"
          fi
          ;;
        rocky|almalinux)
          os_type="rocky"
          ;;
        fedora)
          os_type="rocky"  # 使用 Rocky 脚本（dnf 兼容）
          ;;
        *)
          log "检测到 Linux 发行版: $ID，尝试根据包管理器自动选择..."
          if command -v apt-get >/dev/null 2>&1; then
            os_type="ubuntu"
          elif command -v dnf >/dev/null 2>&1; then
            os_type="rocky"
          elif command -v yum >/dev/null 2>&1; then
            os_type="centos"
          else
            echo "错误：无法识别的 Linux 发行版，且未找到 apt-get/dnf/yum" >&2
            exit 1
          fi
          ;;
      esac
    else
      # 旧系统可能没有 /etc/os-release，尝试其他方式
      if command -v apt-get >/dev/null 2>&1; then
        os_type="ubuntu"
      elif command -v dnf >/dev/null 2>&1; then
        os_type="rocky"
      elif command -v yum >/dev/null 2>&1; then
        os_type="centos"
      else
        echo "错误：无法检测 Linux 发行版类型" >&2
        exit 1
      fi
    fi
  else
    echo "错误：不支持的操作系统类型: $OSTYPE" >&2
    exit 1
  fi

  echo "$os_type"
}

# 主逻辑
main() {
  log "检测操作系统..."
  OS_TYPE=$(detect_os)
  log "检测到操作系统: $OS_TYPE"

  PREPARE_SCRIPT="${SCRIPT_DIR}/prepare_${OS_TYPE}.sh"

  if [ ! -f "$PREPARE_SCRIPT" ]; then
    echo "错误：找不到对应的准备脚本: $PREPARE_SCRIPT" >&2
    exit 1
  fi

  log "执行准备脚本: $PREPARE_SCRIPT"

  if [ -x "$PREPARE_SCRIPT" ]; then
    exec "$PREPARE_SCRIPT"
  else
    exec bash "$PREPARE_SCRIPT"
  fi
}

main

