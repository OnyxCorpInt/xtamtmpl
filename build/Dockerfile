#
# Stage 1: build go code
FROM golang:1.11 as go-builder

# Setup module root
RUN mkdir -p /src/xtamtmpl
WORKDIR /src/xtamtmpl

# Install dependencies
COPY go.mod .
COPY go.sum .
RUN go mod download

# Build app
COPY . .
RUN go install ./cmd/xtamtmpl

# Stage 2: get CA certificates (needed to authenticate XTAM servers with valid certs)
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