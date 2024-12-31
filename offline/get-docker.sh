#! /bin/bash
. .env

function get_docker() {
  if [ ! -e $FILES_OUTPUT/$DOCKER_TAR ]; then
    echo "==> Download $DOCKER_URL"
    for i in {1..3}; do
      curl --location --show-error --fail --output $FILES_OUTPUT/$DOCKER_TAR $DOCKER_URL && return
      echo "curl failed. Attempt=$i"
    done
    echo "Download failed, exit : $DOCKER_URL"
    exit 1
  else
    echo "==> Skip $DOCKER_URL"
  fi
}

function remove_docker() {
  rm -f $FILES_OUTPUT/$DOCKER_TAR
}