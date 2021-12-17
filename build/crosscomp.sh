#!/bin/sh

GOOS=solaris
GOOS=openbsd
GOOS=linux

GOARCH=386
GOARCH=amd64

echo "compiling for $GOOS $GOARCH"

# go build -v github.com\zew\https-server -o https-server-new
go build -v  .\cmd\server\main.go
rm       -f   go-questionnaire-new
mv      main  go-questionnaire-new

echo "success"

sleep 10


