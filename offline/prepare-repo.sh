#! /bin/bash
. .env
. common.sh
. check-env.sh

function create_repo() {
  get_system_version
  if [ "$VERSION_MAJOR" == "CentOS-7" ]; then
    echo "==> createrepo"
    createrepo $REPO_OUTPUT/$ID-$VERSION_ID/$ARCH || exit 1
    echo "create-repo done."
  elif [ "$VERSION_MAJOR" == "CentOS-8" ]; then
    echo "==> createrepo"
    createrepo $REPO_OUTPUT/$ID-$VERSION_ID/$ARCH || exit 1
    echo "create-repo done."
  elif [ "$VERSION_MAJOR" == "kylin-V10" ]; then
    echo "==> createrepo"
    createrepo $REPO_OUTPUT/$ID-$VERSION_ID/$ARCH || exit 1
    echo "create-repo done."
  elif [ "$VERSION_MAJOR" == "Ubuntu-22.04" ]; then
    echo "===> Creating repo"
    pushd $REPO_OUTPUT/$ID-$VERSION_ID/$ARCH || exit 1
    apt-ftparchive sources . > Sources && gzip -c9 Sources > Sources.gz
    apt-ftparchive packages . > Packages && gzip -c9 Packages > Packages.gz
    apt-ftparchive contents . > Contents-$ARCH && gzip -c9 Contents-$ARCH > Contents-$ARCH.gz
    apt-ftparchive release . > Release
    popd
    echo "Done."
  elif [ "$VERSION_MAJOR" == "Ubuntu-24.04" ]; then
    echo "===> Creating repo"
    pushd $REPO_OUTPUT/$ID-$VERSION_ID/$ARCH || exit 1
    apt-ftparchive sources . > Sources && gzip -c9 Sources > Sources.gz
    apt-ftparchive packages . > Packages && gzip -c9 Packages > Packages.gz
    apt-ftparchive contents . > Contents-$ARCH && gzip -c9 Contents-$ARCH > Contents-$ARCH.gz
    apt-ftparchive release . > Release
    popd
    echo "Done."
  elif [ "$VERSION_MAJOR" == "uos-20" ]; then
    echo "==> createrepo"
    createrepo $REPO_OUTPUT/$ID-$VERSION_ID/$ARCH || exit 1
    echo "create-repo done."
  else
    echo "===>Unsupported System Version: $VERSION_MAJOR"
  fi

}

function configure_local_repo() {
  get_system_version
  if [ "$VERSION_MAJOR" == "CentOS-7" ] || [ "$VERSION_MAJOR" == "CentOS-8" ] || [ "$VERSION_MAJOR" == "kylin-V10" ] || [ "$VERSION_MAJOR" == "uos-20" ]; then
    cat <<EOF | sudo tee /etc/yum.repos.d/local.list > /dev/null
[local-mirror]
name=Local Mirror
baseurl=file://$REPO_OUTPUT/$ID-$VERSION_ID/$ARCH
enabled=1
gpgcheck=0
EOF
  elif [ "$VERSION_MAJOR" == "Ubuntu-22.04" ] || [ "$VERSION_MAJOR" == "Ubuntu-24.04" ] || [ "$VERSION_MAJOR" == "Debian-12" ]; then
    cat <<EOF | sudo tee /etc/apt/sources.list.d/local.list > /dev/null
deb [trusted=yes] file://$REPO_OUTPUT/$ID-$VERSION_ID/$ARCH/ /
EOF
  fi
}


function configure_repo() {
  get_system_version
  current_ip=$(get_main_ip)
  if [ "$VERSION_MAJOR" == "CentOS-7" ] || [ "$VERSION_MAJOR" == "CentOS-8" ] || [ "$VERSION_MAJOR" == "kylin-V10" ] || [ "$VERSION_MAJOR" == "uos-20" ]; then
    echo "Creating local repository....."
    cat <<EOF | tee /etc/yum.repos.d/mymirror.repo > /dev/null
[local-mirror]
name=Local Mirror
baseurl=http://$current_ip:$FILE_SERVER_PORT/repository/$ID-$VERSION_ID/$ARCH
enabled=1
gpgcheck=0
EOF
    echo "Creating done."
  elif [ "$VERSION_MAJOR" == "Ubuntu-22.04" ] || [ "$VERSION_MAJOR" == "Ubuntu-24.04" ] || [ "$VERSION_MAJOR" == "Debian-12" ]; then
    echo "Creating local repository....."
    cat <<EOF | tee /etc/apt/sources.list.d/mymirror.list > /dev/null
deb [trusted=yes] http://$current_ip:$FILE_SERVER_PORT/repository/$ID-$VERSION_ID/$ARCH/ /
EOF
    echo "Creating done."
  else
    echo "===>Unsupported System Version: $VERSION_MAJOR"
  fi
}


function remove_repo() {
  get_system_version
  current_ip=$(get_main_ip)
  if [ "$VERSION_MAJOR" == "CentOS-7" ] || [ "$VERSION_MAJOR" == "CentOS-8" ] || [ "$VERSION_MAJOR" == "kylin-V10" ] || [ "$VERSION_MAJOR" == "uos-20" ]; then
    rm -f /etc/apt/sources.list.d/mymirror.list
  elif [ "$VERSION_MAJOR" == "Ubuntu-22.04" ] || [ "$VERSION_MAJOR" == "Ubuntu-24.04" ] || [ "$VERSION_MAJOR" == "Debian-12" ]; then
    rm -f /etc/apt/sources.list.d/mymirror.list
  fi
}