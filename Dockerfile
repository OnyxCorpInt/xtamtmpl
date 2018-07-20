#
# Stage 1: build go code
FROM golang as go-builder

# Setup main package
RUN mkdir /go/src/xtamtmpl
WORKDIR /go/src/xtamtmpl

# Copy over packages
COPY . .

# Install dependencies
RUN go get -u github.com/golang/dep/cmd/dep
RUN dep ensure

# Build app
RUN go install

#
# Stage 2: main entrypoint
FROM debian
ADD VERSION .
COPY --from=go-builder /go/bin/xtamtmpl /usr/local/bin/xtamtmpl
ENTRYPOINT ["xtamtmpl"]