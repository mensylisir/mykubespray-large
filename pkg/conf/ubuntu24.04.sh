# ===================================================================
# Stage 1: The Builder
# - Installs tools, downloads DEBs, and creates the ISO
# ===================================================================
FROM ubuntu:22.04 as builder

# --- Arguments and Environment Variables ---
ARG TARGETARCH=amd64
ARG OS_RELEASE=jammy # 'jammy' is the codename for Ubuntu 22.04
ENV DEBIAN_FRONTEND=noninteractive
ARG BUILD_TOOLS="apt-transport-https software-properties-common ca-certificates curl wget gnupg dpkg-dev genisoimage dirmngr"
ARG REPO_DIR_NAME=ubuntu2204-${TARGETARCH}-debs
ARG ISO_NAME=ubuntu2204-${TARGETARCH}-repo.iso

WORKDIR /tmp

# --- Prepare the build environment in a single layer ---
RUN apt-get update && \
    apt-get install -y --no-install-recommends ${BUILD_TOOLS} && \
    # Add Docker's official GPG key (the modern, secure way)
    mkdir -p /etc/apt/keyrings && \
    curl -fsSL "https://download.docker.com/linux/ubuntu/gpg" | gpg --dearmor -o /etc/apt/keyrings/docker.gpg && \
    # Add the Docker repository
    echo "deb [arch=${TARGETARCH} signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu ${OS_RELEASE} stable" > /etc/apt/sources.list.d/docker.list && \
    apt-get update && \
    apt-get clean

# --- Download packages and create the repository ---
COPY ubuntu22.04.packages .
RUN \
    # Generate a list of URLs for all packages and their dependencies
    apt-get install --yes --reinstall --print-uris $(xargs < ubuntu22.04.packages) \
    | awk -F"'" '{print $2}' \
    | grep -E '^https?://' \
    | sort -u > packages.urls && \
    # Create the repository directory and download all files
    mkdir -p ${REPO_DIR_NAME} && \
    wget -q -x -P ${REPO_DIR_NAME} -i packages.urls && \
    # Create the repository metadata
    cd ${REPO_DIR_NAME} && \
    # The deb files from wget are inside host-named directories, move them up
    find . -type f -name "*.deb" -exec mv {} . \; && \
    # Clean up empty directories left by wget -x
    find . -type d -empty -delete && \
    dpkg-scanpackages . /dev/null | gzip -9c > ./Packages.gz && \
    cd /tmp

# --- Create the ISO with an optimized internal structure ---
RUN cd ${REPO_DIR_NAME} && \
    mkisofs -r -o /tmp/${ISO_NAME} .

# ===================================================================
# Stage 2: The Final Artifact
# - A minimal stage to hold the final ISO file for easy extraction.
# ===================================================================
FROM scratch

# --- Copy the final ISO from the builder stage ---
ARG ISO_NAME=ubuntu2204-amd64-repo.iso
COPY --from=builder /tmp/${ISO_NAME} /${ISO_NAME}