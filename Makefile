SHELL := /bin/bash -x
APP := shylock
VERSION := `cat VERSION`

# Support binary builds
PLATFORMS := linux darwin freebsd

all: build

clean:
	rm -fr build 
	echo $(PLATFORMS)
	@- $(foreach PLAT,$(PLATFORMS), \
		mkdir -p build/$(PLAT) \
		)

deps:
	# Someday switch to vgo once it works with the code
	# go get -u golang.org/x/vgo

	# vgo build

	go get -u github.com/golang/dep/cmd/dep

vendor: deps
	dep ensure


build: clean 

	for plat in $(PLATFORMS); do \
		echo "Building $$plat ..." ; \
		GOARCH=amd64 GOOS=$$plat go build -ldflags "-s -w" -o build/$$plat/$(APP) cmd/shylock/main.go ; \
	done


test: 
	go test ./...

test-integration: 
	go test ./... --tags=integration

package: 
	./packaging/render.py

package-clean:
	rm -rf pkg-build

package-deb: package-clean
	go get -u github.com/mh-cbon/go-bin-deb
	go-bin-deb generate -a amd64 --version $(VERSION) -w pkg-build/deb/amd64/ -o build/$(APP).deb -f build/deb/deb.json
	sudo dpkg -i build/$(APP).deb

package-rpm: package-clean
	go get -u github.com/mh-cbon/go-bin-rpm
	go-bin-rpm generate -a amd64 --version $(VERSION) -b pkg-build/rpm/amd64/ -o build/$(APP).rpm -f build/rpm/rpm.json
	sudo rpm -i build/$(APP).rpm

