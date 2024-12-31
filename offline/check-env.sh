#! /bin/bash
. .env

function check_docker() {
  if command -v docker &> /dev/null; then
    docker_version=$(docker --version | awk '{print $3}'| cut -d ',' -f1)
    echo "Docker ${docker_version} Already installed"
  else
    echo "Error: Docker is not installed or not in PATH. Aborting."
    exit 1
  fi
}