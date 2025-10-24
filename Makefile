CURRENT_DIR=$(shell pwd)
DIST_DIR=${CURRENT_DIR}/dist

.PHONY: clean lint test

tinyapp:
	go build -o ${DIST_DIR}/tinyapp -v ${CURRENT_DIR}/cmd/main.go

tinyapp-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 make tinyapp

$(GOPATH)/bin/golangci-lint:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b `go env GOPATH`/bin v1.42.0

.PHONY: lint
lint: $(GOPATH)/bin/golangci-lint
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.42.0
	wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.42.0
	golangci-lint run -E revive -E goimports -E gocritic -E goconst -E gofmt -E ifshort -E misspell -E whitespace --concurrency 4 --timeout 15m

test:
	go test $(shell go list ./... | grep -v /vendor/ | grep "/controllers\|/gateway\|/server") -race -short -v -coverprofile=coverage.out

clean:
	rm -rf ${CURRENT_DIR}/dist