FROM 172.30.1.13:18093/kylin/kylin-server-10-sp3b23-x86_64-20230324:b23-arm64 as kylinv10sp3-arm64
ENV TARGETARCH=arm64
ENV OS=kylin
ENV OS_VERSION=v10sp3
ARG BUILD_TOOLS="dnf-utils createrepo genisoimage "
ARG DIR=${OS}${OS_VERSION}-${TARGETARCH}-rpms

WORKDIR /tmp

# RUN sed -i '/\[ks10-adv-addons\]/{:a;n;/enabled = 0/s/enabled = 0/enabled = 1/;ba}' /etc/yum.repos.d/kylin_x86_64.repo
RUN yum -q -y update
RUN yum install -q -y ${BUILD_TOOLS}
# RUN yum-config-manager --add-repo https://mirrors.aliyun.com/docker-ce/linux/centos/docker-ce.repo
RUN yum makecache

COPY kylinv10sp3.packages .
RUN mkdir -p $DIR

RUN cd $DIR && sort -u /tmp/kylinv10sp3.packages | xargs repotrack && cd - \
    && createrepo -d ${DIR} \
    && mkisofs -r -o ${DIR}.iso ${DIR}
