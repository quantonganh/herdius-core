#!/bin/bash

protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/gogo/protobuf/protobuf --go_out=. accounts/protobuf/stream.proto
protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/gogo/protobuf/protobuf --go_out=. hbi/protobuf/service.proto