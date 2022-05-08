#!/bin/bash

if [[ "$1" == "--help" ]]; then
    echo ""
    echo "Copies the required library from the github.com/Traceableai/goagent/filter/traceable package to dynamically link it in runtime."
    echo ""
    echo "Usage:"
    echo "        $0 [<destination_dir>]"
    echo ""
    exit 0
fi

set -e

DST_DIR=${1:-.}

if [[ -f $DST_DIR/libtraceable.so ]]; then
    echo "Library already present in \"$DST_DIR/libtraceable.so\", remove it before attempting to copy it."
    exit 1
fi

if [[ ! -f go.mod ]]; then
    echo "go.mod file not found"
    exit 1
fi

set +e
IS_GOAGENT_REQUIRED=$(cat go.mod | grep -ic "github.com/Traceableai/goagent v")
set -e
if [[ "$IS_GOAGENT_REQUIRED" == "0" ]]; then
    echo "github.com/Traceableai/goagent isn't a required package in go.mod"
    exit 1
fi

# For compatibility issues we aim to use the same libtraceable library as the one
# used in the build. To achieve that we look for the location of the library and
# resolved by go modules. The obtained info from `go mod download -json <pkg>` looks
# like this:
#
# {
#        "Path": "honnef.co/go/tools",
#        "Version": "v0.0.0-20190523083050-ea95bdfd59fc",
#        "Info": "/my-go-root/pkg/mod/cache/download/honnef.co/go/tools/@v/v0.0.0-20190523083050-ea95bdfd59fc.info",
#        "GoMod": "/my-go-root/pkg/mod/cache/download/honnef.co/go/tools/@v/v0.0.0-20190523083050-ea95bdfd59fc.mod",
#        "Zip": "/my-go-root/pkg/mod/cache/download/honnef.co/go/tools/@v/v0.0.0-20190523083050-ea95bdfd59fc.zip",
#        "Dir": "/my-go-root/pkg/mod/honnef.co/go/tools@v0.0.0-20190523083050-ea95bdfd59fc",
#        "Sum": "h1:/hemPrYIhOhy8zYrNj+069zDB68us2sMGsfkFJO0iZs=",
#        "GoModSum": "h1:rf3lG4BRIbNafJWhAfAdb/ePZxsR/4RtNHQocxwk9r4="
# }
#
# Hence we deal with this output to find the "Dir" value.
#
GOAGENT_DIR=$(go mod download -json github.com/Traceableai/goagent \
    | head -7 | tail -1 \
    | awk -F\" '{print $4}')

mkdir -p $DST_DIR

IS_ALPINE="0"
if [[ -f /etc/os-release ]]; then
    set +e
    IS_ALPINE=$(cat /etc/os-release | grep "NAME=" | grep -ic "Alpine")
    set -e
fi

if [[ "$IS_ALPINE" == "0" ]]; then # not alpine
    cp ${GOAGENT_DIR}/filter/traceable/libs/linux_$(go env GOARCH)/libtraceable.so $DST_DIR/libtraceable.so
    cp ${GOAGENT_DIR}/filter/traceable/libtraceable.h $DST_DIR/libtraceable.h
    echo "Linux library successfuly copied to $DST_DIR/libtraceable.so"
else
    cp ${GOAGENT_DIR}/filter/traceable/libs/linux_$(go env GOARCH)-alpine/libtraceable.so $DST_DIR/libtraceable.so
    cp ${GOAGENT_DIR}/filter/traceable/libtraceable.h $DST_DIR/libtraceable.h
    echo "Alpine library successfuly copied to $DST_DIR/libtraceable.so"
fi
