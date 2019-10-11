#!/bin/bash

pushd `dirname $0`/.. > /dev/null

if [[ "$1" == "all" ]]
then
  go fmt ./... && \
    GOOS=linux go build -o ./manifer_linux -i ./cmd/manifer && \
    GOOS=darwin go build -o ./manifer_darwin -i ./cmd/manifer
else
  go fmt ./... && \
    go build -o ./manifer -i ./cmd/manifer
fi
result=$?

popd > /dev/null

exit $result