#!/bin/sh

# call from app root via
# .\build\build.bat


# argument -o only works from within local directory
# cd ..\cmd\server\
# go build -v -race   -o "server.exe"


# go build -v -race ./cmd/server/main.go
#    would only work for a *single* file in ./cmd/server/
cd ./cmd/server/
go build -v -race

cd ..
cd ..
rm       -f      server.exe
mv      ./cmd/server/server.exe  ./server.exe

# packing the executable
# upx -1 server.exe

./server.exe

# echo "success"
# sleep 10

