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
RUN go install ./cmd/xtamtmpl

# Stage 2: get CA certificates (needed for XTAM)
FROM debian as ca-certs
RUN apt-get update
RUN apt-get install -y ca-certificates

#
# Stage 2: main entrypoint
FROM debian
ADD VERSION .
COPY --from=go-builder /go/bin/xtamtmpl /usr/local/bin/xtamtmpl
COPY --from=ca-certs /usr/share/ca-certificates /usr/share/ca-certificates
COPY --from=ca-certs /etc/ssl/certs /etc/ssl/certs
ENTRYPOINT ["xtamtmpl"]