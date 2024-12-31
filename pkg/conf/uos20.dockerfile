FROM registry.dev.rdev.tech:18093/uos-server-base/uos-server-20-1070e:latest-amd64 as uos20-amd64
ARG TARGETARCH
ENV OS=uos
ENV OS_VERSION=20
ARG BUILD_TOOLS="dnf-utils createrepo genisoimage "
ARG DIR=${OS}${OS_VERSION}-${TARGETARCH}-rpms

WORKDIR /tmp

RUN yum -q -y update
RUN yum install -q -y ${BUILD_TOOLS}
# RUN yum-config-manager --add-repo https://mirrors.aliyun.com/docker-ce/linux/centos/docker-ce.repo
RUN yum makecache

COPY uos20.packages .
RUN mkdir -p $DIR

RUN cd $DIR && sort -u /tmp/uos20.packages | xargs repotrack && cd - \
    && createrepo -d ${DIR} \
    && mkisofs -r -o ${DIR}.iso ${DIR}
