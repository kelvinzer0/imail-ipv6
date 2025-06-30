#!/bin/bash
set -e

# curl -fsSL  https://raw.githubusercontent.com/midoks/imail/master/scripts/install.sh | sh

check_go_environment() {
	if test ! -x "$(command -v go)"; then
		printf "\e[1;31mmissing go running environment\e[0m\n"
		exit 1
	fi;
}

load_vars() {
	OS=$(uname | tr '[:upper:]' '[:lower:]')

	VERSION=$(get_latest_release "kelvinzer0/imail-ipv6")

	TARGET_DIR="/usr/local/imail"
};

get_latest_release() {
    curl -sL "https://api.github.com/repos/$1/releases/latest" | grep '"tag_name":' | cut -d'"' -f4;
}

get_arch() {
	echo "package main
import (
	\"fmt\"
	\"runtime\"
)
func main() { fmt.Println(runtime.GOARCH) }" > /tmp/go_arch.go

	ARCH=$(go run /tmp/go_arch.go);
}

get_download_url() {
	DOWNLOAD_URL="https://github.com/kelvinzer0/imail-ipv6/releases/download/$VERSION/imail_$(echo $VERSION | sed 's/^v//')_${OS}_${ARCH}.tar.gz";
}

# download file
download_file() {
    url="${1}"
    destination="${2}"

    printf "Fetching ${url} \n\n"

    if test -x "$(command -v curl)"; then
        code=$(curl --connect-timeout 15 -w '%{http_code}' -L "${url}" -o "${destination}")
    elif test -x "$(command -v wget)"; then
        code=$(wget -t2 -T15 -O "${destination}" --server-response "${url}" 2>&1 | awk '/^  HTTP/{print $2}' | tail -1)
    else
        printf "\e[1;31mNeither curl nor wget was available to perform http requests.\e[0m\n"
        exit 1
    fi

    if [ "${code}" != 200 ]; then
        printf "\e[1;31mRequest failed with code %s\e[0m\n" $code
        exit 1
    else 
	    printf "\n\e[1;33mDownload succeeded\e[0m\n"
    fi;
}


main() {
	check_go_environment

	load_vars

	local_tar_gz=""
	if [ -n "$1" ]; then
		local_tar_gz="$1"
	fi

	if [ -f "$TARGET_DIR/imail" ]; then
		printf "\n\e[1;32mImail is already installed in %s. Skipping download and extraction.\e[0m\n" "$TARGET_DIR"
	else
		if [ -z "$local_tar_gz" ]; then
			get_arch
			get_download_url
			DOWNLOAD_FILE="$(mktemp).tar.gz"
			download_file "$DOWNLOAD_URL" "$DOWNLOAD_FILE"
		else
			DOWNLOAD_FILE="$local_tar_gz"
			if [ ! -f "$DOWNLOAD_FILE" ]; then
				printf "\e[1;31mError: Local tar.gz file not found: %s\e[0m\n" "$DOWNLOAD_FILE"
				exit 1
			fi
			printf "\n\e[1;33mUsing local tar.gz file: %s\e[0m\n" "$DOWNLOAD_FILE"
		fi

		if [ ! -d "$TARGET_DIR" ]; then
			mkdir -p "$TARGET_DIR"
		fi

		tar -C "$TARGET_DIR" -zxf "$DOWNLOAD_FILE"
		if [ -z "$local_tar_gz" ]; then
			rm -rf "$DOWNLOAD_FILE"
		fi
	fi

	# Copy docs directory if it exists in the extracted tarball
	if [ -d "$TARGET_DIR/docs" ]; then
		cp -r "$TARGET_DIR/docs" "$TARGET_DIR/docs_backup"
	fi

	# Create systemd service file
	sed "s:{APP_PATH}:${TARGET_DIR}:g" "$TARGET_DIR/scripts/init.d/imail.service.tpl" > /etc/systemd/system/imail.service

	systemctl daemon-reload
	systemctl restart imail

	"$TARGET_DIR/imail" -v	

	printf "\n\e[1;32mImail installation complete!\e[0m\n"
	printf "\e[1;32mPlease complete the initial setup by visiting http://localhost:1080/install in your web browser.\e[0m\n"
};

main "$@" || exit 1