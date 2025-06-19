# ===================================================================
# Stage 1: The Builder
# - Installs tools, downloads RPMs, and creates the ISO
# ===================================================================
FROM registry.dev.rdev.tech:18093/uos-server-base/uos-server-20-1070e:latest-arm64 as builder

# --- Arguments and Environment Variables ---
ARG TARGETARCH=amd64
ENV OS=uos
ENV OS_VERSION=20
ARG BUILD_TOOLS="dnf-utils createrepo genisoimage"
ARG REPO_DIR_NAME=${OS}${OS_VERSION}-${TARGETARCH}-rpms
ARG ISO_NAME=${OS}${OS_VERSION}-${TARGETARCH}-repo.iso

WORKDIR /tmp

# --- Prepare the build environment in a single layer ---
RUN yum -q -y update && \
    yum install -q -y ${BUILD_TOOLS} && \
    # 如果需要外部仓库, 在这里添加。例如:
    # yum-config-manager --add-repo https://mirrors.aliyun.com/docker-ce/linux/centos/docker-ce.repo
    yum makecache && \
    yum clean all

# --- Download packages and create the repository ---
COPY uos20.packages .
RUN mkdir -p ${REPO_DIR_NAME} && \
    # 使用 xargs -a 代替 cat | xargs，更简洁
    repotrack -p ${REPO_DIR_NAME} -a uos20.packages && \
    # 在 RPM 目录内创建仓库元数据
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
ARG ISO_NAME=uos20-amd64-repo.iso
COPY --from=builder /tmp/${ISO_NAME} /${ISO_NAME}