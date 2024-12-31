#! /bin/bash
. .env
. common.sh
. check-env.sh

function download_repo() {
  get_system_version
  if [ ! -e $REPO_OUTPUT/$ID-$VERSION_ID/$ARCH ]; then
    mkdir -p $REPO_OUTPUT/$ID-$VERSION_ID/$ARCH
  fi
  if [ "$VERSION_MAJOR" == "CentOS-7" ]; then
    packages=$(cat dnf.list | grep -v "^#" | sort | uniq)
    repotrack -p $REPO_OUTPUT/$ID-$VERSION_ID/$ARCH $packages || {
          echo "Download error"
          exit 1
  }
  elif [ "$VERSION_MAJOR" == "CentOS-8" ]; then
    packages=$(cat dnf.list | grep -v "^#" | sort | uniq)
    repotrack -p $REPO_OUTPUT/$ID-$VERSION_ID/$ARCH $packages || {
          echo "Download error"
          exit 1
  }
  elif [ "$VERSION_MAJOR" == "kylin-V10" ]; then
    packages=$(cat kylin-v10.list | grep -v "^#" | sort | uniq)
    repotrack --downloaddir $REPO_OUTPUT/$ID-$VERSION_ID/$ARCH $packages || {
          echo "Download error"
          exit 1
  }
  elif [ "$VERSION_MAJOR" == "Ubuntu-22.04" ]; then
    packages=$(cat ubuntu-22.04.list | grep -v "^#" | sort | uniq)
    echo "===> Install Repository"
    sudo apt update
    sudo apt install -y apt-transport-https ca-certificates curl gnupg lsb-release apt-utils
    echo "===> Update apt cache"
    sudo apt update
    echo "===> Resolving dependencies"
    DEPS=$(apt-cache depends --recurse --no-recommends --no-suggests --no-conflicts --no-breaks --no-replaces --no-enhances --no-pre-depends $packages | grep "^\w" | sort | uniq)
    echo "===> Downloading packages: " $packages $DEPS
    cd $REPO_OUTPUT/$ID-$VERSION_ID/$ARCH && apt download $packages $DEPS && cd -
  elif [ "$VERSION_MAJOR" == "Ubuntu-24.04" ]; then
    packages=$(cat apt.list | grep -v "^#" | sort | uniq)
    echo "===> Install Repository"
    sudo apt update
    sudo apt install -y apt-transport-https ca-certificates curl gnupg lsb-release apt-utils
    echo "===> Update apt cache"
    sudo apt update
    echo "===> Resolving dependencies"
    DEPS=$(apt-cache depends --recurse --no-recommends --no-suggests --no-conflicts --no-breaks --no-replaces --no-enhances --no-pre-depends $packages | grep "^\w" | sort | uniq)
    echo "===> Downloading packages: " $packages $DEPS
    cd $REPO_OUTPUT/$ID-$VERSION_ID/$ARCH && apt download $packages $DEPS && cd -
  elif [ "$VERSION_MAJOR" == "Debian-12" ]; then
    packages=$(cat apt.list | grep -v "^#" | sort | uniq)
    echo "===> Install Repository"
    sudo apt update
    sudo apt install -y apt-transport-https ca-certificates curl gnupg lsb-release apt-utils
    echo "===> Update apt cache"
    sudo apt update
    echo "===> Resolving dependencies"
    DEPS=$(apt-cache depends --recurse --no-recommends --no-suggests --no-conflicts --no-breaks --no-replaces --no-enhances --no-pre-depends $packages | grep "^\w" | sort | uniq)
    echo "===> Downloading packages: " $packages $DEPS
    cd $REPO_OUTPUT/$ID-$VERSION_ID/$ARCH && apt download $packages $DEPS && cd -
  elif [ "$VERSION_MAJOR" == "uos-20" ]; then
    packages=$(cat uos-20.list | grep -v "^#" | sort | uniq)
    repotrack --downloaddir $REPO_OUTPUT/$ID-$VERSION_ID/$ARCH $packages || {
          echo "Download error"
          exit 1
  }
  else
    echo "===> Unsupported System Version: $VERSION_MAJOR"
  fi
}

function download_multi_repo() {
  get_system_version
  for iarch in ${MULTI_ARCH}; do
    if [ "$iarch" == "amd64" ]; then
      iarch="x86_64"
    fi
    if [ ! -e $REPO_OUTPUT/$ID-$VERSION_ID/$iarch ]; then
     mkdir -p $REPO_OUTPUT/$ID-$VERSION_ID/$iarch
    fi
    if [ "$VERSION_MAJOR" == "CentOS-7" ]; then
      packages=$(cat dnf.list | grep -v "^#" | sort | uniq)
      repotrack -a $iarch -p $REPO_OUTPUT/$ID-$VERSION_ID/$iarch $packages || {
            echo "Download error"
            exit 1
    }
    elif [ "$VERSION_MAJOR" == "CentOS-8" ]; then
      packages=$(cat dnf.list | grep -v "^#" | sort | uniq)
      repotrack -a $iarch -p $REPO_OUTPUT/$ID-$VERSION_ID/$iarch $packages || {
            echo "Download error"
            exit 1
    }
    elif [ "$VERSION_MAJOR" == "kylin-V10" ]; then
      packages=$(cat dnf.list | grep -v "^#" | sort | uniq)
      repotrack -a $iarch -p $REPO_OUTPUT/$ID-$VERSION_ID/$iarch $packages || {
            echo "Download error"
            exit 1
    }
    elif [ "$VERSION_MAJOR" == "Ubuntu-22.04" ]; then
      packages=$(cat ubuntu-22.04.list | grep -v "^#" | sort | uniq)
      echo "===> Install Repository"
      sudo apt update
      sudo apt install -y apt-transport-https ca-certificates curl gnupg lsb-release apt-utils
      echo "===> Update apt cache"
      sudo apt update
      echo "===> Resolving dependencies"
      DEPS=$(apt-cache depends --recurse --no-recommends --no-suggests --no-conflicts --no-breaks --no-replaces --no-enhances --no-pre-depends $packages | grep "^\w" | sort | uniq)
      DEPS=$(echo $DEPS | awk -v iarch="$iarch" '{for(i=1;i<=NF;i++) printf "%s:%s ", $i, iarch}')
      echo "===> Downloading packages: " $packages $DEPS
      cd $REPO_OUTPUT/$ID-$VERSION_ID/$iarch && apt download $packages $DEPS && cd -
    elif [ "$VERSION_MAJOR" == "Ubuntu-24.04" ]; then
      packages=$(cat apt.list | grep -v "^#" | sort | uniq)
      echo "===> Install Repository"
      sudo apt update
      sudo apt install -y apt-transport-https ca-certificates curl gnupg lsb-release apt-utils
      echo "===> Update apt cache"
      sudo apt update
      echo "===> Resolving dependencies"
      DEPS=$(apt-cache depends --recurse --no-recommends --no-suggests --no-conflicts --no-breaks --no-replaces --no-enhances --no-pre-depends $packages | grep "^\w" | sort | uniq)
      DEPS=$(echo $DEPS | awk -v iarch="$iarch" '{for(i=1;i<=NF;i++) printf "%s:%s ", $i, iarch}')
      echo "===> Downloading packages: " $packages $DEPS
      cd $REPO_OUTPUT/$ID-$VERSION_ID/$iarch && apt download $packages $DEPS && cd -
    elif [ "$VERSION_MAJOR" == "Debian-12" ]; then
      packages=$(cat apt.list | grep -v "^#" | sort | uniq)
      echo "===> Install Repository"
      sudo apt update
      sudo apt install -y apt-transport-https ca-certificates curl gnupg lsb-release apt-utils
      echo "===> Update apt cache"
      sudo apt update
      echo "===> Resolving dependencies"
      DEPS=$(apt-cache depends --recurse --no-recommends --no-suggests --no-conflicts --no-breaks --no-replaces --no-enhances --no-pre-depends $packages | grep "^\w" | sort | uniq)
      DEPS=$(echo $DEPS | awk -v iarch="$iarch" '{for(i=1;i<=NF;i++) printf "%s:%s ", $i, iarch}')
      echo "===> Downloading packages: " $packages $DEPS
      cd $REPO_OUTPUT/$ID-$VERSION_ID/$iarch && apt download $packages $DEPS && cd -
    elif [ "$VERSION_MAJOR" == "uos-20" ]; then
      packages=$(cat uos-20.list | grep -v "^#" | sort | uniq)
      repotrack -a $iarch -p $REPO_OUTPUT/$ID-$VERSION_ID/$iarch $packages || {
            echo "Download error"
            exit 1
    }
    fi
  done
}

function remove_repo() {
  get_system_version
  rm -rf $REPO_OUTPUT/$ID-$VERSION_ID/$ARCH
}