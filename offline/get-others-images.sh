#! /bin/bash
. .env
. common.sh

function get_other_images() {
  other_images=$(cat images.list | sed "s/#.*$//g" | sort -u)
  for image in $other_images; do
      image=$(expand_image_repo $image)
      get_image $image
  done
}