#! /bin/bash
. .env
. common.sh

#function get_image() {
#  if [ ! -e $IMAGES_OUTPUT ]; then
#    mkdir -p $IMAGES_OUTPUT
#  fi
#  image=$1
#  tar_name=$(echo ${image##*/} | sed s/:/-/g).tar
#  if [ ! -e $IMAGES_OUTPUT/tar_name ]; then
#    echo "===> Pull $image"
#    docker pull $image || exit 1
#    docker save -o $IMAGES_OUTPUT/$tar_name $image || exit 1
#  else
#    echo "==> Skip $image"
#  fi
#}

function download_images() {
  if [ ! -e $IMAGE_LIST ]; then
    generate_list
  fi
  images=$(cat $IMAGE_LIST)
  for image in $images; do
    get_image $image  
  done
}

function remove_images() {
  remove_registry
  rm -rf $IMAGES_OUTPUT
}