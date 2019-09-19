#!/bin/bash

pushd `dirname $0`/.. > /dev/null

go list github.com/golang/mock/mockgen >/dev/null || go install github.com/golang/mock/mockgen

function mock {
	set -e
	path=$1

	# prepend mock to filename
	mockfile=$(echo $path | sed -E 's/(.*)\/([^/]+)\/([^/]+).go/\1\/\2\/mock_\3.go/')

	# extract folder name
	packagename=$(echo $path  | sed -E 's/(.*)\/([^/]+)\/([^/]+).go/\2/')

	$GOPATH/bin/mockgen -source $path -destination $mockfile -package $packagename

	# remove 'x "."' import - https://github.com/golang/mock/issues/230
	sed -i "s/*x\./*/g; s/ x\./ /g; s/\[\]x\./[]/g; /x \".\"/d" $mockfile

	# remove mock files for packages with no interfaces (unused gomock import)
	find . -name mock_*.go | xargs -I{} bash -c "grep -q NewMock {} || rm {}"
}
export -f mock

# regenerate mocks for all interfaces
find . -name *.go | grep -v test | grep -v mock | xargs -I{} bash -c "mock {}"

go fmt ./...
go vet ./...

go test $@ ./cmd/... ./pkg/...
result=$?

# clean up generated files
find . -name mock_*.go -exec rm {} \;

popd > /dev/null

exit $result
