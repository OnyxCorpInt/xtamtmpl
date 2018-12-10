PKGS := $(shell go list ./... | grep -v -P "vendor")
TAG?=$(or ${VERSION},$(shell cat VERSION))
.PHONY: lint image install xtamtmpl

xtamtmpl:
	rm -f xtamtmpl
	go build ./cmd/xtamtmpl

install: xtamtmpl
	go install ./cmd/xtamtmpl

image:
	docker build -t xtamtmpl:$(TAG) -f ./build/Dockerfile .

lint:
	go install golang.org/x/lint/golint
	go vet $(PKGS)
	golint $(PKGS)
