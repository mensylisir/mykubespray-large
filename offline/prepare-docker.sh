#! /bin/bash
. .env
. common.sh
. check-env.sh

function install_docker() {
  echo "Installing Docker ${DOCKER_VERSION}..."
  tar -zxvf $FILES_OUTPUT/$DOCKER_TAR -C .tmp
  cp .tmp/docker/* /usr/local/bin
  echo "Cleaning up..."
  rm -rf /tmp/docker
  generate_service
  systemctl daemon-reload
  systemctl enable docker
  systemctl start docker
  sleep 3
  res=$(systemctl is-active docker)
  if [ $res == "active" ];  then
    echo "Docker started success"
  else
    echo "Docker start failed"
    exit 1
  fi
}

function generate_service() {
  if [ ! -e "/etc/systemd/system/docker.service" ]; then
    echo "Creating docker.service file..."
    cat <<EOF | sudo tee /etc/systemd/system/docker.service > /dev/null
[Unit]
Description=Docker Application Container Engine
Documentation=https://docs.docker.com
After=network-online.target docker.socket
Wants=network-online.target

[Service]
ExecStart=/usr/local/bin/dockerd
ExecReload=/bin/kill -s HUP \$MAINPID
TimeoutSec=0
RestartSec=2
Restart=always
StartLimitBurst=3
StartLimitInterval=60s
LimitNOFILE=infinity
LimitNPROC=infinity
LimitCORE=infinity
Delegate=yes
KillMode=process

[Install]
WantedBy=multi-user.target
EOF
  fi
}


function remove_docker() {
  rm -f /usr/local/bin/docker-proxy
  rm -f /usr/local/bin/containerd
  rm -f /usr/local/bin/dockerd
  rm -f /usr/local/bin/docker
  rm -f /usr/local/bin/containerd-shim-runc-v2
  rm -f /usr/local/bin/ctr
  rm -f /usr/local/bin/docker-init
  rm -f /usr/local/bin/runc
  rm -f /usr/local/bin/containerd-shim
  rm -f /etc/systemd/system/docker.service
  rm -rf /etc/systemd/system/docker.service.d
  rm -rf /etc/docker
}

function configure_proxy() {
  conf_dir="/etc/systemd/system/docker.service.d"
  if [ ! -e $conf_dir ]; then
    mkdir -p $conf_dir
  fi
  current_ip=$(get_main_ip)
  docker_no_proxy="${DOCKER_NO_PROXY},$current_ip:$REGISTRY_PORT"
  tee $conf_dir/http-proxy.conf > /dev/null <<EOF
[Service]
Environment="HTTP_PROXY=${DOCKER_HTTP_PROXY}"
Environment="HTTPS_PROXY=${DOCKER_HTTPS_PROXY}"
Environment="NO_PROXY=${docker_no_proxy}"
EOF
  systemctl daemon-reload
  systemctl restart docker
  res=$(systemctl is-active docker)
  if [ $res == "active" ];  then
    echo "Docker started success"
  else
    echo "Docker start failed"
    exit 1
  fi
}
