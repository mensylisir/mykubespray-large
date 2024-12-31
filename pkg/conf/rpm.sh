#!/usr/bin/env bash

BUILD_TOOLS="yum-utils createrepo mkisofs epel-release"
DIR=iso
yum install -q -y ${BUILD_TOOLS}
yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
yum makecache
sort -u packages.list | xargs repotrack -p ${DIR}
createrepo -d ${DIR}
mkisofs -r -o ${DIR}.iso ${DIR}
