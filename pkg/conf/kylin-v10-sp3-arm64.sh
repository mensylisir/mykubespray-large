FROM 172.30.1.13:18093/kylin/kylin-server-10-sp3b23-x86_64-20230324:b23-arm64 as builder

# --- Arguments and Environment Variables ---
ARG TARGETARCH=amd64
ENV OS=kylin
ENV OS_VERSION=v10sp3
# 注意: epel-release 可能有兼容性风险，如果不需要请移除
ARG BUILD_TOOLS="yum-utils createrepo genisoimage epel-release"
ARG REPO_DIR_NAME=${OS}${OS_VERSION}-${TARGETARCH}-rpms
ARG ISO_NAME=${OS}${OS_VERSION}-${TARGETARCH}-repo.iso

WORKDIR /tmp

# --- Prepare the build environment in a single layer ---
RUN yum install -q -y ${BUILD_TOOLS} && \
    # 注意: 这个是CentOS的Docker源，在Kylin上可能不兼容，请优先使用官方适配源
    yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo && \
    yum makecache && \
    yum clean all

# --- Download packages and create the repository ---
COPY kylinv10sp3.packages .
RUN mkdir -p ${REPO_DIR_NAME} && \
    # 使用 xargs -a 代替 cat | xargs
    repotrack -p ${REPO_DIR_NAME} $(xargs < kylinv10sp3.packages) && \
    createrepo -d ${REPO_DIR_NAME}

# --- Create the ISO with an optimized internal structure ---
RUN cd ${REPO_DIR_NAME} && \
    mkisofs -r -o /tmp/${ISO_NAME} .

# ===================================================================
# Stage 2: The Final Artifact
# - A minimal stage to hold the final ISO file for easy extraction.
# ===================================================================
FROM scratch

# --- Copy the final ISO from the builder stage ---
ARG ISO_NAME=kylinv10sp3-amd64-repo.iso
COPY --from=builder /tmp/${ISO_NAME} /${ISO_NAME}