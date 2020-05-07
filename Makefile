# Target: app/os/arch
TARGETS = watcher/linux/amd64 sender/linux/amd64

VERSION = $(shell ./scripts/version.sh)
export BUILD_PACKAGE = github.com/void616/gm.mint.sender
export BUILD_VERSION = $(BUILD_PACKAGE)/internal/version.version=$(VERSION)
export BUILD_TAGS =

OUTPUT_DIR = ./build/bin
BRANCH = $(shell git rev-parse --abbrev-ref HEAD)

ifndef GOPATH
	$(error GOPATH is undefined)
endif
ifndef GM_DOCKER_PREFIX
	$(error GM_DOCKER_PREFIX is undefined)
endif


split = $(word $2,$(subst /, ,$1))

.PHONY: build

all: deps test build image push

deps:
	@echo "Ensure you've installed Protobuf compiler https://github.com/protocolbuffers/protobuf/releases"

test: gen
	go test ./...

build: clean gen $(TARGETS)

clean:
	rm -rf ./build/bin/* | true
	mkdir -p ./build/bin | true

gen:
	go generate ./...

$(TARGETS):
	@{ \
	export APP=$(call split,$@,1) ;\
	export BUILD_APP=cmd/$$APP/main.go ;\
	export BUILD_OS=$(call split,$@,2) ;\
	export BUILD_ARCH=$(call split,$@,3) ;\
	if [ "$$BUILD_OS" == "windows" ]; then APPEXT=.exe; fi ;\
	export BUILD_OUTFILE=$${APP}-$${BUILD_ARCH}$${APPEXT} ;\
	export BUILD_OUTDIR=$(OUTPUT_DIR)/$${APP}-$${BUILD_OS} ;\
	\
	echo "Building $$BUILD_APP ($${BUILD_OS}/$${BUILD_ARCH})" ;\
	./scripts/gobuild.sh ;\
	}

image:
	docker build -t $(GM_DOCKER_PREFIX)/mintsender-watcher:$(BRANCH) -f ./build/dockerfile-watcher-linux-amd64 .
	docker build -t $(GM_DOCKER_PREFIX)/mintsender-sender:$(BRANCH) -f ./build/dockerfile-sender-linux-amd64 .
	docker push $(GM_DOCKER_PREFIX)/mintsender-watcher:$(BRANCH)
	docker push $(GM_DOCKER_PREFIX)/mintsender-sender:$(BRANCH)