#!/bin/bash -exu

ROOT=${PWD}
GOPATH=${PWD}/go
PATH=${GOPATH}/bin:$PATH

function main() {
  mkdir -p "${GOPATH}/src/github.com/pivotal-cf"
  ln -s "${ROOT}/azure-blobstore-resource" "${GOPATH}/src/github.com/pivotal-cf/azure-blobstore-resource"
  pushd "${GOPATH}/src/github.com/pivotal-cf/azure-blobstore-resource" > /dev/null
    ./scripts/build
  popd > /dev/null

  cp -R ${ROOT}/azure-blobstore-resource/* "${ROOT}/workspace"
}

main
