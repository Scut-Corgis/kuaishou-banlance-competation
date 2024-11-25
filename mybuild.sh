#!/bin/bash

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

cd ..
# 编译B
cd B

go build -o bin/B pkg/main/main.go



rm -rf output
mkdir output
chmod +x bin/start.sh && chmod +x bin/stop.sh
cp -r bin output/

