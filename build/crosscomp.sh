#!/bin/sh

# export is needed
export GOOS=solaris
export GOOS=openbsd
export GOOS=linux

export GOARCH=386
export GOARCH=amd64

echo "ENV is $GOOS $GOARCH"
echo "effective golang ENV is"
go env GOOS GOARCH
sleep 1

# go build -v -race ./cmd/server/main.go
#    would only work for a *single* file in ./cmd/server/
cd ./cmd/server/
echo "we are in"
pwd
go build -v  -o server-new


echo "success"

sleep 10


