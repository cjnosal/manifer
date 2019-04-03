#!/bin/bash

SETUP=$1
pushd `dirname $0`/.. > /dev/null

flagsrepo=github.com/bosh-dep-forks/go-flags
flagsbranch=cli-patches

boshrepo=github.com/xtreme-conor-nosal/bosh-cli
boshbranch=temp-module-name

function setup() {
	go get $flagsrepo@$flagsbranch
	go get $boshrepo@$boshbranch
}

go fmt ./...

# if packages are missing or --setup requested, go get required branches of forked repos
go list $flagsrepo >/dev/null 2>&1 && go list $boshrepo >/dev/null 2>&1
MISSING=$?
if [[ $SETUP == "--setup" || $MISSING != 0 ]]
then
	setup
fi

go build -o ./manifer -i ./cmd/manifer 

popd > /dev/null