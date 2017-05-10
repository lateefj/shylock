#!/bin/bash
yum update
yum upgrade -y
yum install -y fuse libfuse-dev fakeroot 
yum install -y htop sysstat make httperf wget curl git rpm-build

curl https://storage.googleapis.com/golang/go1.8.1.linux-amd64.tar.gz > $$HOME/go1.8.1.linux-amd64.tar.gz
cd $$HOME
tar xvf go1.8.1.linux-amd64.tar.gz
export GOPATH=$$HOME
export GOROOT=$$GOPATH/go
export PATH=$$PATH:$$GOPATH/go/bin:$$GOPATH/bin

cd /${app_name}/
make package-${package_manager}

