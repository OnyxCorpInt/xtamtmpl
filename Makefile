PKGS := $(shell go list ./... | grep -v -P "vendor")
TAG?=$(or ${VERSION},$(shell cat VERSION))
REGISTRY=$(shell cat REGISTRY)
.PHONY: lint image install xtamtmpl

xtamtmpl:
	rm -f xtamtmpl
	go build ./cmd/xtamtmpl

install: xtamtmpl
	go install ./cmd/xtamtmpl

image:
	docker build -t xtamtmpl:$(TAG) -f ./build/Dockerfile .

tag-registry: image
	docker tag xtamtmpl:$(TAG) $(REGISTRY)/xtamtmpl:$(TAG)

publish: tag-registry
	docker push $(REGISTRY)/xtamtmpl:$(TAG)

lint:
	go install golang.org/x/lint/golint
	go vet $(PKGS)
	golint $(PKGS)
