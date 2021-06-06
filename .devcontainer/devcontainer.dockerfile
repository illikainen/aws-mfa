FROM golang:buster

ENV LANG="C" \
    GO111MODULE="on" \
    GOFLAGS="-mod=readonly" \
    GOPROXY="off" \
    GOSUMDB="sum.golang.org"


RUN apt-get update && \
    apt-get -y dist-upgrade && \
    apt-get -y install awscli bash curl git netcat unzip vim

RUN useradd --create-home --user-group --shell /bin/bash dev
USER dev

# v0.6.11
RUN GOPROXY="https://proxy.golang.org" \
    go install -mod=readonly -v \
    golang.org/x/tools/gopls@0bb7e5c47b1a31f85d4f173edc878a8e049764a5

# v0.1.2
RUN GOPROXY="https://proxy.golang.org" \
    go install -mod=readonly -v \
    golang.org/x/tools/cmd/goimports@c3e30ffe9275c0e89ef0d38049bad16987bd7fb2

# v1.39.0
RUN GOPROXY="https://proxy.golang.org" \
    go install -mod=readonly -v \
    github.com/golangci/golangci-lint/cmd/golangci-lint@9aea4aee1c0d47f74c016c1f0066cb90e2f7d2e8
