#!/bin/bash
#https://gist.github.com/bclinkinbeard/1331790
versionLabel=$1
productName="goDash"
releaseFolder="release/"

rm ${releaseFolder}*.tgz
rm ${releaseFolder}*.zip

rice embed-go

arch=amd64
os=darwin
product=${releaseFolder}${productName}_${versionLabel}_${os}_${arch}
env GOOS=$os GOARCH=$arch go build -ldflags="-s -w -X main.version=${1} -X main.buildstamp=`date -u '+%Y-%m-%d_%I:%M:%S%p'` " -o $product cmd/main.go
tar czfv $product.tgz $product
rm $product

arch=amd64
os=linux
product=${releaseFolder}${productName}_${versionLabel}_${os}_${arch}
env GOOS=$os GOARCH=$arch go build -ldflags="-s -w -X main.version=${1} -X main.buildstamp=`date -u '+%Y-%m-%d_%I:%M:%S%p'` " -o $product cmd/main.go
tar czfv $product.tgz $product
rm $product

arch=386
os=linux
product=${releaseFolder}${productName}_${versionLabel}_${os}_${arch}
env GOOS=$os GOARCH=$arch go build -ldflags="-s -w -X main.version=${1} -X main.buildstamp=`date -u '+%Y-%m-%d_%I:%M:%S%p'` " -o $product cmd/main.go
tar czfv $product.tgz $product
rm $product

arch=arm
os=linux
product=${releaseFolder}${productName}_${versionLabel}_${os}_${arch}
env GOOS=$os GOARCH=$arch go build -ldflags="-s -w -X main.version=${1} -X main.buildstamp=`date -u '+%Y-%m-%d_%I:%M:%S%p'` " -o $product cmd/main.go
tar czfv $product.tgz $product
rm $product

arch=amd64
os=windows
product=${releaseFolder}${productName}_${versionLabel}_${os}_${arch}
env GOOS=$os GOARCH=$arch go build -ldflags="-s -w -X main.version=${1} -X main.buildstamp=`date -u '+%Y-%m-%d_%I:%M:%S%p'` " -o ${product}.exe cmd/main.go
zip -r ${product}.zip ${product}.exe
rm $product.exe

arch=386
os=windows
product=${releaseFolder}${productName}_${versionLabel}_${os}_${arch}
env GOOS=$os GOARCH=$arch go build -ldflags="-s -w -X main.version=${1} -X main.buildstamp=`date -u '+%Y-%m-%d_%I:%M:%S%p'` " -o ${product}.exe cmd/main.go
zip -r ${product}.zip ${product}.exe
rm $product.exe

