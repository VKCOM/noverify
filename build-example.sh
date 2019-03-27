#!/bin/sh
NOW=`date +%Y%m%d%H%M%S`
OS=`uname -n -m`
AFTER_COMMIT=`git rev-parse HEAD`

go build -ldflags "-X 'github.com/VKCOM/noverify/src/cmd.BuildTime=$NOW' -X 'github.com/VKCOM/noverify/src/cmd.BuildOSUname=$OS' -X 'github.com/VKCOM/noverify/src/cmd.BuildCommit=$AFTER_COMMIT'" -o noverify
