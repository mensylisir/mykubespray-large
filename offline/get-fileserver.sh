#! /bin/bash
. .env
. common.sh
. check-env.sh

function make_file_server() {
  check_docker
  docker build -f Dockerfile.fileserver -t $FILE_SERVER_IMAGE .
  if [ $? -eq 0 ];  then
    echo "$FILE_SERVER_IMAGE make success"
  else
    echo "$FILE_SERVER_IMAGE make failed"
  fi
  tar_name=$(echo ${FILE_SERVER_IMAGE##*/} | sed s/:/-/g).tar
  docker save -o $IMAGES_OUTPUT/$ARCH/$tar_name $FILE_SERVER_IMAGE
}