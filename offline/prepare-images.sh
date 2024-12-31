#! /bin/bash
. .env
. check-env.sh
. common.sh


function create_registry() {
  check_docker
  docker_ps=$(docker ps | grep $REGISTRY_NAME)
  if [ -n "$docker_ps" ]; then
      echo "Error: Docker registry already exists"
  else
    tar_name=$(echo ${REGISTRY_IMAGE##*/} | sed s/:/-/g).tar
    echo "===>loading $REGISTRY_IMAGE"
    docker load -i $IMAGES_OUTPUT/$ARCH/$tar_name
    echo "===> starting docker repository"
    docker run -d -p $REGISTRY_PORT:5000 -v /opt/registry:/mnt/registry --restart=always --name $REGISTRY_NAME $REGISTRY_IMAGE
    if [ $? -ne 0 ]; then
      echo "Error: Failed to run Docker Registry container."
      exit 1
    fi
    current_ip=$(get_main_ip)
    echo "Docker Registry is running at http://$current_ip:$REGISTRY_PORT"
  fi
}



function configure_docker() {
  current_ip=$(get_main_ip)
  echo "Configuring Docker client to allow insecure registry..."
  daemon_config="/etc/docker/daemon.json"
  if [ ! -f "$daemon_config" ]; then
    mkdir -p /etc/docker/
    tee /etc/docker/daemon.json > /dev/null <<EOF
{
  "insecure-registries": ["$current_ip:$REGISTRY_PORT"]
}
EOF
  fi
  echo "Restarting Docker service to apply configuration..."
  systemctl daemon-reload
  systemctl restart docker
  sleep 3
#  docker_ps=$(docker ps | grep $REGISTRY_NAME)
#  if [ -z "$docker_ps" ]; then
#      echo "Error: Docker service restart failed. Please check Docker logs."
#      exit 1
#  fi
  echo "Docker service restarted successfully."
}

function pushed_images() {
  ls -la $IMAGES_OUTPUT/$ARCH/*.tar | awk '{print $NF}' > $IMAGES_OUTPUT/$ARCH/image-tar.list
  images_tar=$(cat $IMAGES_OUTPUT/$ARCH/image-tar.list)
  for image_tar_name in $images_tar; do
    res=$(docker load -i $image_tar_name)
    tmp_image=$(echo $res | awk -F"Loaded image:" '{print $NF}' | tr -d ' ')
    tmp2_image=$(expand_image_repo $tmp_image)
    current_ip=$(get_main_ip)
    new_image="$current_ip:$REGISTRY_PORT/$(echo ${tmp2_image#*/})"
    docker tag $tmp2_image $new_image
    docker push $new_image
  done
}

function pushed_multi_images() {
  for iarch in ${MULTI_ARCH}; do
    ls -la $IMAGES_OUTPUT/$iarch/*.tar | awk '{print $NF}' > $IMAGES_OUTPUT/$iarch/image-tar.list
    images_tar=$(cat $IMAGES_OUTPUT/$iarch/image-tar.list)
    for image_tar_name in $images_tar; do
      res=$(docker load -i $image_tar_name)
      tmp_image=$(echo $res | awk -F"Loaded image:" '{print $NF}' | tr -d ' ')
      tmp2_image=$(expand_image_repo $tmp_image)
      current_ip=$(get_main_ip)
      new_image="$current_ip:$REGISTRY_PORT/$(echo ${tmp2_image#*/})"
      docker tag $tmp2_image $new_image
      docker push $new_image
    done
  done
}

function remove_registry() {
  docker_ps=$(docker ps | grep $REGISTRY_NAME)
  if [ -n "$docker_ps" ]; then
      docker rm -f $REGISTRY_NAME
      exit 0
  fi
}