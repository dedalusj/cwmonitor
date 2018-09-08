OUT := cwmonitor
GIT_VERSION := $(shell git describe --always || echo "dev")
BUILD_TIME := $(shell date -u +"%Y%m%dT%H%M%S")
VERSION := $(GIT_VERSION)-$(BUILD_TIME)
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
	docker build -t "dedalusj/${OUT}:${VERSION}" .

push: docker
	docker push dedalusj/${OUT}:${VERSION}

.PHONY: run build vet lint