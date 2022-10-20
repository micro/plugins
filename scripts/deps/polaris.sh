#!/bin/bash

OS="linux"
ARCH="amd64"
ORG="polarismesh"
REPO="polaris"
PORT="8090"

echo "Downloading Polaris"

# Find latest release
package_url=$(curl -s "https://api.github.com/repos/${ORG}/${REPO}/releases/latest" |
	grep "https://github.com/polarismesh/polaris/releases/download" | grep "standalone" | grep "$OS" | grep "$ARCH" |
	cut -d : -f 2,3 |
	tr -d '\"[:space:]')

echo "Downloading from '${package_url}'"

# Downlaod ladtest release
wget -q "${package_url}"


# Unzip package
unzip -o "*.linux.amd64.zip"

package_name=$(find . -maxdepth 1 -name "polaris*" -type d)
if [ ! -d "${package_name}" ]; then
	echo "${package_name} doesn't exist"
	exit 1
fi

# Change dir to polaris dir
pushd "${package_name}" || exit 1

# Run install script
echo "Running install script"
bash install.sh

# Give it some safety margin to startup
sleep 5

# Check if port open
if ! nc -z "127.0.0.1" "${PORT}"; then
	echo "Failed to find a service running on $PORT"

	# Echo netstat
	netstat -tulpn | grep -i polaris
fi

popd || exit 1

# Export Address
export POLARIS_ADDR="127.0.0.1:8091"

printf "\nPolaris installed successfully on ${POLARIS_ADDR}\n"

