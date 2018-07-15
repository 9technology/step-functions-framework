PROJECT = stepworkerlauncher
OUT_DIR = temp
REGISTRY = your.dockerregistry.com
PWD = $(shell pwd)

deps:
	echo "Installing dependencies ..."
	cp -r ../common/my9awsgo src/
.PHONY: deps

build: deps
	export GOPATH=$(GOPATH):$(PWD); cd src/main; GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o stepworkerlauncher runstatemachine.go runstepworker.go time.go stepworkerlauncher.go
	cp src/main/stepworkerlauncher temp/
	cp src/main/*.js temp/
	cd temp; zip stepworkerlauncher *
.PHONY: build

# Use this make target to clean sources locally on desktop  ***
clean:
	rm -f src/main/$(PROJECT)
	rm -f temp/$(PROJECT)
	rm -rf src/my9awsgo
	rm -f temp/*.js
	rm -f temp/*.zip
.PHONY: clean
