#! /bin/bash
. .env
. common.sh
. check-env.sh

function create_file_server() {
  check_docker
  docker_ps=$(docker ps | grep "$FILE_SERVER_NAME")
  if [ -n "$docker_ps" ]; then
      echo "Error: Docker $FILE_SERVER_NAME already exists"
  else
    tar_name=$(echo ${FILE_SERVER_IMAGE##*/} | sed s/:/-/g).tar
    echo "===>loading $FILE_SERVER_IMAGE"
    docker load -i $IMAGES_OUTPUT/$ARCH/$tar_name
    echo "===> starting fileserver"
    docker run -d -p $FILE_SERVER_PORT:80 -v $REPO_OUTPUT:/usr/share/nginx/html/repository -v $FILES_OUTPUT:/usr/share/nginx/html/files --restart=always --name $FILE_SERVER_NAME $FILE_SERVER_IMAGE
    if [ $? -ne 0 ]; then
      echo "Error: Failed to run Docker fileserver container."
      exit 1
    fi
    current_ip=$(get_main_ip)
    echo "Docker fileserver is running at http://$current_ip:$FILE_SERVER_PORT"
  fi

}

function remove_file_server() {
  docker_ps=$(docker ps | grep $REGISTRY_NAME)
  if [ -n "$docker_ps" ]; then
      docker rm -f $FILE_SERVER_NAME
      exit 0
  fi
}