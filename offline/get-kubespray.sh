#! /bin/bash
. .env
. check-env.sh
. common.sh

function create_kubespray_cache_dir() {
  if [ ! -e $KUBESPRAY_CACHE ]; then
    /bin/mkdir -p $KUBESPRAY_CACHE
  fi
}

function create_kubespray_temp_dir() {
  if [ ! -e $KUBESPRAY_DIR ]; then
    /bin/mkdir -p $KUBESPRAY_DIR
  fi
}


function remove_kubespray_cache_dir() {
  if [ -e $KUBESPRAY_CACHE ]; then
    /bin/rm -rf $KUBESPRAY_CACHE
  fi
}

function remove_kubespray_tmp_dir() {
  if [ -e $KUBESPRAY_DIR ]; then
    /bin/rm -rf $KUBESPRAY_DIR
  fi
}


function get_kubespray() {
  echo "===> Download ${KUBESPRAY_TARBALL}"
  curl -SL https://codeload.github.com/kubernetes-sigs/kubespray/tar.gz/refs/tags/v$KUBESPRAY_VERSION >$KUBESPRAY_DIR/${KUBESPRAY_TARBALL} || exit 1
  echo "===> Download success."
}

function extract_kubespray() {
  echo "===> Extract ${KUBESPRAY_TARBALL}"
  tar -zxvf $KUBESPRAY_DIR/${KUBESPRAY_TARBALL} -C $KUBESPRAY_CACHE  || exit 1
  echo "===> Extract done."
}


function download_kubespray() {
    create_kubespray_cache_dir
    create_kubespray_temp_dir
    get_kubespray
    extract_kubespray
}

function remove_kubespray() {
  remove_kubespray_cache_dir
  remove_kubespray_tmp_dir
}


function generate_kubespray() {
  cat <<EOF | sudo tee Dockerfile.kubespray > /dev/null
FROM ubuntu:24.04
MAINTAINER liminggang@163.com
USER root
ENV LANG=C.UTF-8 \
    DEBIAN_FRONTEND=noninteractive \
    PYTHONDONTWRITEBYTECODE=1
WORKDIR /kubespray
SHELL ["/bin/bash", "-c"]
RUN apt-get update -q \
       && apt-get install -y \
       sshpass \
       python3.12 \
       python3.12-venv \
       python3-pip \
       vim \
       rsync \
       ansible \
       apt-utils \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* /var/log/*

RUN pip config set global.index-url https://mirrors.ustc.edu.cn/pypi/web/simple

COPY .cache/kubespray-$KUBESPRAY_VERSION/requirements.txt /kubespray/requirements.txt

RUN python3 -m venv /opt/myenv
ENV PATH="/opt/myenv/bin:$PATH"
RUN pip install --upgrade pip setuptools wheel
RUN pip install --no-compile --no-cache-dir -r requirements.txt \
    && find /usr -type d -name '*__pycache__' -prune -exec rm -rf {} \;
RUN ansible-galaxy collection install community.general --force
RUN ansible-galaxy collection install ansible.posix --force
RUN ansible-galaxy collection install ansible.network --force
RUN ansible-galaxy collection install community.kubernetes --force
RUN ansible-galaxy collection install community.docker --force
RUN ansible-galaxy collection install containers.podman --force
EOF
}

function make_kubespray() {
  check_docker
  generate_kubespray
  kubespray_name=$KUBESPRAY_IMAGE:v$KUBESPRAY_VERSION
  docker build -f Dockerfile.kubespray -t $KUBESPRAY_IMAGE:v$KUBESPRAY_VERSION .
  if [ $? -eq 0 ];  then
    echo "$kubespray_name make success"
  else
    echo "$kubespray_name make failed"
  fi
  tar_name=$(echo ${kubespray_name##*/} | sed s/:/-/g).tar
  docker save -o $IMAGES_OUTPUT/$ARCH/$tar_name $kubespray_name
}
