#!/bin/sh
set -eu

REPO="SomNOG/somnog"
BINARY="somnog"
INSTALL_DIR="/usr/local/bin"

get_arch() {
  arch=$(uname -m)
  case "$arch" in
    x86_64|amd64) echo "amd64" ;;
    arm64|aarch64) echo "arm64" ;;
    *) echo "Unsupported architecture: $arch" >&2; exit 1 ;;
  esac
}

get_os() {
  os=$(uname -s | tr '[:upper:]' '[:lower:]')
  case "$os" in
    linux) echo "linux" ;;
    darwin) echo "darwin" ;;
    *) echo "Unsupported OS: $os" >&2; exit 1 ;;
  esac
}

main() {
  os=$(get_os)
  arch=$(get_arch)

  echo "Detected: ${os}/${arch}"

  tag=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')

  if [ -z "$tag" ]; then
    echo "Error: could not determine latest release" >&2
    exit 1
  fi

  echo "Installing somnog ${tag}..."

  url="https://github.com/${REPO}/releases/download/${tag}/${BINARY}-${os}-${arch}"

  tmpdir=$(mktemp -d)
  trap 'rm -rf "$tmpdir"' EXIT

  curl -fsSL -o "${tmpdir}/${BINARY}" "$url"
  chmod +x "${tmpdir}/${BINARY}"

  if [ -w "$INSTALL_DIR" ]; then
    mv "${tmpdir}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
  else
    echo "Need sudo to install to ${INSTALL_DIR}"
    sudo mv "${tmpdir}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
  fi

  echo "somnog installed to ${INSTALL_DIR}/${BINARY}"
  echo ""
  somnog version
  echo ""
  echo "Get started:"
  echo "  somnog new my-app"
  echo "  cd my-app"
  echo "  somnog start"
}

main