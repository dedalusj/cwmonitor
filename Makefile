OUT := cwmonitor
GIT_REVISION := $(shell git rev-parse HEAD || echo "dev")
GIT_VERSION := $(shell git describe --always --tags || echo "dev")
BUILD_TIME := $(shell date -u +"%Y%m%dT%H%M%SZ")
VERSION := $(GIT_VERSION)-$(BUILD_TIME)
BUILD_NUMBER := local
DOCKER_IMAGE := dedalusj/$(OUT)
DOCKER_TAG := $(GIT_VERSION)
PKG_LIST := $(shell go list ./... | grep -v /vendor/)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/)

all: run

build:
	CGO_ENABLED=0 go build -i -v -o ${OUT} -ldflags="-X main.version=${VERSION}" .

build-linux:
	CGO_ENABLED=0 GOOS=linux go build -i -v -o ${OUT} -ldflags="-X main.version=${VERSION}" .

test:
	@go test -short -v ${PKG_LIST}

test-all:
	@go test -v ${PKG_LIST}

vet:
	@go vet ${PKG_LIST}

lint:
	@for file in ${GO_FILES} ;  do \
		golint $$file ; \
	done

run: build
	./${OUT}

clean:
	-@rm ${OUT}

coverage:
	@go test -coverprofile=.coverage.out -covermode=atomic ${PKG_LIST}
	@go tool cover -html .coverage.out

docker: build-linux
	docker build \
	  --build-arg created=${BUILD_TIME} \
	  --build-arg version=${GIT_VERSION} \
	  --build-arg revision=${GIT_REVISION} \
	  --build-arg build_number=${BUILD_NUMBER} \
	  -t "${DOCKER_IMAGE}:${DOCKER_TAG}" .

push: docker
	docker tag "${DOCKER_IMAGE}:${DOCKER_TAG}" "${DOCKER_IMAGE}:latest"
	docker push "${DOCKER_IMAGE}:${DOCKER_TAG}"
	docker push "${DOCKER_IMAGE}:latest"

.PHONY: run build vet lint