#!/bin/sh

if [ x$GOPATH == x ]; then 
export GOPATH=`pwd`
fi

if [ $# -eq 0 ]; then

echo "empty target"
exit 1

fi

Folder=$1

githash=`git rev-parse HEAD`
build=`date -u '+%Y-%m-%d_%H:%M:%S'`
buildA=`date -u '+%Y_%m_%d_%H_%M_%S'`

commitDate=`git log --pretty=format:"%h" -1`
headName=`git rev-parse --abbrev-ref HEAD`
gitTagName=`git describe --abbrev=0 --tags`
gitBranchName=`git symbolic-ref --short -q HEAD`

echo $githash $build $commitDate $headName $gitTagName


BaseFlag=" -X static.BuildDate=$build -X static.BuildVersion=$githash -X static.BuildName=$gitBranchName "


Flag=$BaseFlag

if [ x$gitTagName != x ]; then

Flag="$Flag -X static.Version=$gitTagName"

fi

go build         -a          -ldflags "$Flag" -tags=jsoniter  -o "${Folder}_linux_amd64" $Folder

if [ $? -ne 0 ]; then
exit $?
fi
