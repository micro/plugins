#!/bin/bash

OS="linux"
ARCH="amd64"
ORG="polarismesh"
REPO="polaris"
PORT="8090"

echo "Downloading Polaris"

# Find latest release
package_name=$(curl -s "https://api.github.com/repos/${ORG}/${REPO}/releases/latest" |
	grep "https://github.com/polarismesh/polaris/releases/download" | grep "standalone" | grep "$OS" | grep "$ARCH" |
	cut -d : -f 2,3 |
	tr -d \")

# Downlaod ladtest release
wget -qi "${package_name}"

# Unzip package
unzip "*.linux.amd64.zip"
if [ ! -d "${package_name}" ]; then
	echo "${package_name} doesn't exist"
	exit 1
fi

# Change dir to polaris dir
pushd "${package_name}" || exit 1

# Run install script
bash install.sh
sleep 5

# Check if port open
if ! nc -z "127.0.0.1" "${PORT}"; then
	echo "Failed to find a service running on $PORT"

	# Echo netstat
	netstat -tulpn | grep -i polaris
fi

# Find Polaris PIDs
export DEP_PIDS=($(pgrep polaris))

popd || exit 1

# Export Address
export POLARIS_ADDR="127.0.0.1:8091"

echo "Polaris installed successfully on ${POLARIS_ADDR}"

