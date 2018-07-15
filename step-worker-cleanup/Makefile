PROJECT = stepworkercleanup
PROJECT_PATH = step-worker-cleanup
OUT_DIR = temp
REGISTRY = your.dockerregistry.com
PWD = $(shell pwd)
GOLANG_BUILDER = your.dockerregistry.com/tools/golang-build-docker

DEF_PROJECT = default
DEF_ENV = prod
DEF_REGION = ap-southeast-2
DEF_MODE = cleanup_state_machine
DEF_CONFBUCKET = your-step-functions-config
DEF_CONFBUCKETKEY = yourprojects
DEF_RESULT = FAIL::possible-timeout-or-unknown-failure

deps:
	echo "Installing dependencies ..."
	cp -r ../common/my9awsgo src/
.PHONY: deps

build: deps
	export GOPATH=$(GOPATH):$(PWD); \
	cd src/main; GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
	go build  -ldflags \
	"-X main.def_project=$(DEF_PROJECT) -X main.def_env=$(DEF_ENV) -X main.def_region=$(DEF_REGION) -X main.def_mode=$(DEF_MODE) -X main.def_confbucket=$(DEF_CONFBUCKET) -X main.def_confbucketkey=$(DEF_CONFBUCKETKEY) -X main.def_result=$(DEF_RESULT)" \
	 -o stepworkercleanup cleanup.go time.go stepworkercleanup.go
	cp src/main/stepworkercleanup temp/
	cp src/main/*.js temp/
	cd temp; zip stepworkercleanup *
.PHONY: build


# Use this make target to clean sources locally on desktop  ***
clean:
	rm -f src/main/$(PROJECT)
	rm -f temp/$(PROJECT)
	rm -rf src/my9awsgo
	rm -f temp/*.js
	rm -f temp/*.zip
.PHONY: clean
