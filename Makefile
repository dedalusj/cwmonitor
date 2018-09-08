OUT := cwmonitor
GIT_VERSION := $(shell git describe --always || echo "dev")
BUILD_TIME := $(shell date -u +"%Y%m%d%H%M%S")
VERSION := $(GIT_VERSION)-$(BUILD_TIME)
DOCKER_IMAGE := dedalusj/$(OUT)
DOCKER_TAG := $(GIT_VERSION)
PKG_LIST := $(shell go list ./... | grep -v /vendor/)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/)

all: run

build:
	CGO_ENABLED=0 go build -i -v -o ${OUT} -ldflags="-X main.version=${VERSION}" .

test:
	@go test -short -v ${PKG_LIST}

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
	@go test -short -coverprofile=.coverage.out -covermode=atomic ${PKG_LIST}
	@go tool cover -html .coverage.out

docker:
	docker build \
	  --build-arg "version=${VERSION}" \
	  --build-arg "git_version=${GIT_VERSION}" \
	  --build-arg "build_time=${BUILD_TIME}" \
	  -t "${DOCKER_IMAGE}:${DOCKER_TAG}" .

push: docker
	docker tag "${DOCKER_IMAGE}:${DOCKER_TAG}" "${DOCKER_IMAGE}:latest"
	docker push "${DOCKER_IMAGE}:${DOCKER_TAG}"
	docker push "${DOCKER_IMAGE}:latest"

.PHONY: run build vet lint