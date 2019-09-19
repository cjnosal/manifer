#!/bin/bash

pushd `dirname $0`/.. > /dev/null

go fmt ./...

go build -o ./manifer -i ./cmd/manifer
result=$?

popd > /dev/null

exit $result