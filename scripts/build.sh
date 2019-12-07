#!/bin/bash

pushd `dirname $0`/.. > /dev/null

if [[ "$1" == "all" ]]
then
  go fmt ./... && \
    GOOS=linux go build -ldflags="-s -w" -o ./manifer_linux -i ./cmd/manifer && \
    GOOS=darwin go build -ldflags="-s -w" -o ./manifer_darwin -i ./cmd/manifer && \
    GOOS=windows go build -ldflags="-s -w" -o ./manifer_windows.exe -i ./cmd/manifer
else
  go fmt ./... && \
    go build -o ./manifer -i ./cmd/manifer
fi
result=$?

popd > /dev/null

exit $result
