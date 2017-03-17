#!/bin/bash

set -ex

rm -rf bin
mkdir bin

go get github.com/reiver/go-porterstemmer

for app in restapi
do

  GOOS=linux GOARCH=amd64 go build -o bin/${app}  src/${app}.go
  upx -9 bin/${app}

  GOOS=windows GOARCH=amd64 go build -o bin/${app}.exe src/${app}.go
  upx -9 bin/${app}.exe

  GOOS=darwin GOARCH=amd64 go build -o bin/${app}.app src/${app}.go
  upx -9 bin/${app}.app

done
