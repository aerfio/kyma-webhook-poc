#!/usr/bin/env bash

set -e
set -o pipefail

readonly CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly KUBEVAL_VERSION="0.15.0"

host::os() {
  local host_os
  case "$(uname -s)" in
    Darwin)
      host_os=darwin
      ;;
    Linux)
      host_os=linux
      ;;
    *)
      >&2 echo -e "Unsupported host OS. Must be Linux or Mac OS X."
      exit 1
      ;;
  esac
  echo "${host_os}"
}


echo "${os}"

host::createDirIfNotExists(){
  DIR="${CURRENT_DIR}/bin"
  if [ ! -d "$DIR" ]; then
    mkdir -p "${CURRENT_DIR}/bin"
  fi
}

host::testIfFileExists(){
  FILE="${CURRENT_DIR}/bin/kubeval"
  if [ ! -e "${FILE}" ]
  then
    echo "ok"
  else
    echo "nok"
  fi
}

host::createDirIfNotExists

os=$(host::os)

FILE="${CURRENT_DIR}/bin/kubeval"
if [ ! -e "${FILE}" ]
then
  wget "https://github.com/instrumenta/kubeval/releases/download/${KUBEVAL_VERSION}/kubeval-${os}-amd64.tar.gz"
  tar xf "${CURRENT_DIR}/kubeval-${os}-amd64.tar.gz" "kubeval"
  mv "${CURRENT_DIR}/kubeval" "${CURRENT_DIR}/bin/kubeval"
  rm "${CURRENT_DIR}/kubeval-${os}-amd64.tar.gz"
fi

# add all helm configurations
helm template webhook "${CURRENT_DIR}/charts/webhook" | "${CURRENT_DIR}/bin/kubeval" --strict