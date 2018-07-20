TAG?=$(or ${VERSION},$(shell cat VERSION))
.PHONY: lint image install xtamtmpl

xtamtmpl:
	rm -f xtamtmpl
	go build

install: xtamtmpl
	go install xtamtmpl

image:
	docker build -t xtamtmpl:$(TAG) -f Dockerfile .

lint:
	go get golang.org/x/lint/golint
	go vet $(PKGS)
	${GOPATH}/bin/golint $(PKGS)
