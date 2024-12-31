#! /bin/bash
. .env
. common.sh

function get_files() {
    if [ ! -e $FILES_OUTPUT ]; then
      mkdir -p $FILES_OUTPUT
    fi
    url=$1
    file_name=$(echo ${url##*/})
    rdir=$(decide_relative_dir $url)
    if [ -n "$rdir" ]; then
        if [ ! -d $FILES_OUTPUT/$rdir ]; then
            mkdir -p $FILES_OUTPUT/$rdir
        fi
    else
        rdir="."
    fi
    if [ ! -e $FILES_OUTPUT/$rdir/$file_name ]; then
      echo "==> Download $url"
      for i in {1..3}; do
          curl --location --show-error --fail --output $FILES_OUTPUT/$rdir/$file_name $url && return
          echo "curl failed. Attempt=$i"
      done
      echo "Download failed, exit : $url"
      exit 1
    else
      echo "==> Skip $url"
    fi
}

function decide_relative_dir() {
    local url=$1
    local rdir
    rdir=$url
    rdir=$(echo ${rdir#*//})
    rdir=$(echo ${rdir%/*})
    if [ "$url" != "$rdir" ]; then
        echo $rdir
        return
    else
      echo ""
    fi
}

function download_files() {
  if [ ! -e $FILE_LIST ]; then
    generate_list
  fi
  files=$(cat $FILE_LIST)
  for file in $files;do
    get_files $file
  done
}

function remove_files() {
  rm -rf $FILES_OUTPUT
}