#!/bin/bash
crypto="/usr/local/go1.22.0/src/crypto"
sudo rm -rf ${crypto}/sha256
sudo cp -rf ./sha256 ${crypto}/
cat ${crypto}/sha256/sha256.go



cd A 
#设置快手goproxy 镜像
go env -w GO111MODULE="on"
go env -w GOPROXY="https://goproxy.corp.kuaishou.com,direct"
go env -w GOPRIVATE=""
go env -w GONOSUMDB="git.corp.kuaishou.com"

#debug
go version
go env

go build -o bin/A pkg/main/main.go

rm -rf output
mkdir output
chmod +x bin/start.sh && chmod +x bin/stop.sh
cp -r bin output/

# grpc替换
cd ..
grpc="/home/jenkins/go/pkg/mod/google.golang.org/grpc@v1.64.0"
sudo rm -rf ${grpc}/server.go
sudo cp -rf ./others/server.go ${grpc}/server.go

# 编译B
cd B

go build -o bin/B pkg/main/main.go



rm -rf output
mkdir output
chmod +x bin/start.sh && chmod +x bin/stop.sh
cp -r bin output/

