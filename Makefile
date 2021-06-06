PROJECT := aws-mfa
SRC := ./src
GOFLAGS := -mod=readonly

all: fmt build

check:
	GOFLAGS=$(GOFLAGS) golangci-lint run $(SRC)
	GOFLAGS=$(GOFLAGS) go mod verify

fmt:
	GOFLAGS=$(GOFLAGS) go fmt $(SRC)
	GOFLAGS=$(GOFLAGS) goimports -w $(SRC)

build:
	GOFLAGS=$(GOFLAGS) go build -o $(PROJECT) $(SRC)

.PHONY: all check fmt build
