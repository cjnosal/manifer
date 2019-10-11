#!/bin/bash
pushd `dirname $0`/.. > /dev/null

go list github.com/golang/mock/mockgen >/dev/null || \
  (go get github.com/golang/mock/mockgen@v1.2.0 && go install github.com/golang/mock/mockgen)

function mock {
	set -e
	path=$1

	grep -q interface $path || return

	# prepend mock to filename
	mockfile=$(echo $path | sed -E 's/(.*)\/([^/]+)\/([^/]+).go/\1\/\2\/mock_\3.go/')

	# extract folder name
	packagename=$(echo $path  | sed -E 's/(.*)\/([^/]+)\/([^/]+).go/\2/')

	$GOPATH/bin/mockgen -source $path -destination $mockfile -package $packagename

	# remove 'x "."' import - https://github.com/golang/mock/issues/230
	sed -i "s/*x\./*/g; s/ x\./ /g; s/\[\]x\./[]/g; /x \".\"/d" $mockfile
}
export -f mock

# regenerate mocks for all interfaces
find . -name *.go | grep -v test | grep -v mock | xargs -I{} bash -c "mock {}"

# remove mock files for packages with no interfaces (unused gomock import)
find . -name mock_*.go | xargs -I{} bash -c "grep -q NewMock {} || rm {}"

TEST_ARGS="-count=1 ./cmd/... ./lib/... ./pkg/..."
if [ $# -gt 0 ]
then
	if [[ "$1" == "unit" ]]
	then
		TEST_ARGS="./lib/... ./pkg/..."
	elif [[ "$1" == "integration" ]]
	then
		TEST_ARGS="-count=1 ./cmd/..." # count=1 prevents caching
	else
		TEST_ARGS=$@
	fi
fi
echo $TEST_ARGS
go vet ./... && go test $TEST_ARGS
result=$?

# clean up generated files
find . -name mock_*.go -exec rm {} \;

popd > /dev/null

exit $result
