#!/usr/bin/env bash
BUILD_TOOLS="apt-transport-https software-properties-common ca-certificates curl wget gnupg dpkg-dev genisoimage dirmngr"
DIR=iso
apt install -y --no-install-recommends $BUILD_TOOLS
curl -fsSL "https://download.docker.com/linux/ubuntu/gpg" | apt-key add -qq -
echo "deb [arch=$TARGETARCH] https://download.docker.com/linux/ubuntu ${OS_RELEASE} stable" > /etc/apt/sources.list.d/docker.list
apt update -qq
sort -u ubuntu22.04.packages | xargs apt-get install --yes --reinstall --print-uris | awk -F "'" '{print $2}' | grep -v '^$' | sort -u > packages.urls
mkdir -p ${DIR}
wget -q -x -P ${DIR} -i packages.urls
cd ${DIR}
dpkg-scanpackages ./ /dev/null | gzip -9c > ./Packages.gz
cd -
genisoimage -r -o ${DIR}.iso ${DIR}
