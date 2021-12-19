#!/bin/sh

# call from app root via
# .\build\build.bat


# argument -o only works from within local directory
# cd ..\cmd\server\
# go build -v -race   -o "https-server.exe"



# # this always creates main.exe
# go build -v -race ./cmd/server/main.go
# rm       -f      https-server.exe
# mv      main.exe https-server.exe
# # upx -1 https-server.exe
# https-server.exe


# above commands only work for a *single* file in ./cmd/server/
# thus
cd ./cmd/server/

go build -v -race

cd ..
cd ..
rm       -f      https-server.exe
mv      ./cmd/server/server.exe  ./https-server.exe

./https-server.exe

# echo "success"
# sleep 10

