SHELL := /bin/bash
APP := shylock
VERSION := `cat VERSION`

all: build

clean:
	rm -fr build && mkdir -p build/{amd64,darwin}
deps:
	go get -u github.com/kardianos/govendor


vendor: deps
	mkdir -p vendor
	govendor get
	govendor fetch +all
	govendor fmt 

build: clean 
	# Linux build
	GOARCH=amd64 GOOS=linux go build -ldflags "-s -w" -o build/amd64/$(APP)
	# OS X build
	GOARCH=amd64 GOOS=darwin go build -ldflags "-s -w" -o build/darwin/$(APP)

package: 
	./packaging/render.py

package-deb:
	rm -rf pkg-build
	go get -u github.com/mh-cbon/go-bin-deb
	go get -u github.com/mh-cbon/go-bin-rpm
	go-bin-deb generate -a amd64 --version $(VERSION) -w pkg-build/deb/amd64/ -o build/$(APP).deb
	sudo dpkg -i build/$(APP).deb
	# TODO: Need to figure out how to build rpm's on ubuntu
	#go-bin-rpm generate -a amd64 --version $(VERSION -b pkg-build/rpm/amd64/ -o build/$(APP).rpm

