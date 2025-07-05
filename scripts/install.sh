#!/usr/bin/env bash
set -eu

APP=gira
REPO="mklinovsky/gira"
INSTALL_DIR="$HOME/.gira/bin"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
ORANGE='\033[38;2;255;140;0m'
NC='\033[0m' # No Color

print_message() {
    local level=$1
    local message=$2
    local color=""

    case $level in
        info) color="${GREEN}" ;;
        warning) color="${YELLOW}" ;;
        error) color="${RED}" ;;
    esac

    echo -e "${color}${message}${NC}"
}

check_version() {
    local version_to_check=$1
    if command -v $APP >/dev/null 2>&1; then
        installed_version=$($APP --version 2>/dev/null || echo "0.0.0")
        installed_version=${installed_version//v/} # remove potential 'v' prefix

        if [[ "$installed_version" != "$version_to_check" ]]; then
            print_message info "Found installed $APP version: ${YELLOW}$installed_version."
        else
            print_message info "$APP version ${YELLOW}$version_to_check${GREEN} is already installed."
            exit 0
        fi
    fi
}

get_target() {
    local os
    os=$(uname -s | tr '[:upper:]' '[:lower:]')
    local arch
    arch=$(uname -m)

    if [ "$os" == "linux" ]; then
      if [ "$arch" == "x86_64" ]; then
        TARGET="x86_64-unknown-linux-gnu"
      elif [ "$arch" == "aarch64" ]; then
        TARGET="aarch64-unknown-linux-gnu"
      else
        print_message error "Unsupported architecture: $arch"
        exit 1
      fi
    elif [ "$os" == "darwin" ]; then
      if [ "$arch" == "x86_64" ]; then
        TARGET="x86_64-apple-darwin"
      elif [ "$arch" == "arm64" ]; then
        TARGET="aarch64-apple-darwin"
      else
        print_message error "Unsupported architecture: $arch"
        exit 1
      fi
    else
      print_message error "Unsupported operating system: $os"
      exit 1
    fi
    echo $TARGET
}

download_and_install() {
    local version=$1
    local target=$2
    
    local filename="$APP-$version-$target.tar.gz"
    local url="https://github.com/$REPO/releases/download/v$version/$filename"

    print_message info "Downloading ${ORANGE}$APP ${GREEN}version: ${YELLOW}$version${GREEN}..."
    print_message info "URL: $url"

    local tmpdir
    tmpdir=$(mktemp -d)
    
    if ! curl -# -L -o "$tmpdir/$filename" "$url"; then
        print_message error "Download failed. Please check the URL and your connection."
        rm -rf "$tmpdir"
        exit 1
    fi

    tar -xzf "$tmpdir/$filename" -C "$tmpdir"
    
    local binary_in_archive="$APP-$version-$target"
    if [ -f "$tmpdir/$binary_in_archive" ]; then
        mv "$tmpdir/$binary_in_archive" "$INSTALL_DIR/$APP"
    else
        print_message error "Could not find '$binary_in_archive' binary in the downloaded archive."
        ls -lR "$tmpdir"
        rm -rf "$tmpdir"
        exit 1
    fi
    
    rm -rf "$tmpdir"
    print_message info "Installation successful."
}

add_to_path() {
    local config_file=$1
    local command_to_add=$2

    if [[ -w $config_file ]]; then
        echo -e "\n# $APP" >> "$config_file"
        echo "$command_to_add" >> "$config_file"
        print_message info "Successfully added ${ORANGE}$APP${GREEN} to \$PATH in $config_file"
        print_message warning "Please restart your shell or source your config file for the changes to take effect."
    else
        print_message warning "Could not write to $config_file."
        print_message warning "Manually add the directory to your shell's config file:"
        print_message info "  $command_to_add"
    fi
}

mkdir -p "$INSTALL_DIR"

latest_version=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"v?([^"]+)".*/\1/')
if [ -z "$latest_version" ]; then
    print_message error "Could not fetch the latest version."
    exit 1
fi

print_message info "Installing version: ${YELLOW}$latest_version"

check_version "$latest_version"

target=$(get_target)

download_and_install "$latest_version" "$target"

if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    current_shell=$(basename "$SHELL")

    case $current_shell in
        bash)
            if [ -f "$HOME/.bashrc" ]; then config_file="$HOME/.bashrc"; else config_file="$HOME/.bash_profile"; fi
            ;;
        zsh)
            config_file="$HOME/.zshrc"
            ;;
        fish)
            config_file="$HOME/.config/fish/config.fish"
            ;;
        *)
            print_message warning "Could not detect shell. Please add $INSTALL_DIR to your PATH manually."
            exit 0
            ;;
    esac

    if [ -f "$config_file" ]; then
         if [[ "$current_shell" == "fish" ]]; then
            add_to_path "$config_file" "fish_add_path $INSTALL_DIR"
        else
            add_to_path "$config_file" "export PATH=\"$INSTALL_DIR:\$PATH\""
        fi
    else
        print_message warning "Config file $config_file not found."
        print_message warning "Please add $INSTALL_DIR to your PATH manually."
    fi
else
    print_message info "$INSTALL_DIR is already in your PATH."
fi

print_message info "${ORANGE}$APP${GREEN} installation complete!"

