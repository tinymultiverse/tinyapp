CURRENT_DIR=$(shell pwd)
DIST_DIR=${CURRENT_DIR}/dist

.PHONY: clean lint test tinyapp-server tinyapp-server-linux tinyapp-controller tinyapp-controller-linux tinyapp-gateway tinyapp-gateway-linux

clean:
	rm -rf ${CURRENT_DIR}/dist

$(GOPATH)/bin/golangci-lint:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b `go env GOPATH`/bin v1.42.0

.PHONY: lint
lint: $(GOPATH)/bin/golangci-lint
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.42.0
	wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.42.0
	golangci-lint run -E revive -E goimports -E gocritic -E goconst -E gofmt -E ifshort -E misspell -E whitespace --concurrency 4 --timeout 15m
test:
	go test $(shell go list ./... | grep -v /vendor/ | grep -v /test/e2e/) -race -short -v

tinyapp-server:
	go build -mod vendor -o ${DIST_DIR}/tinyapp-server -v ${CURRENT_DIR}/server/cmd/main.go

tinyapp-server-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 make tinyapp-server

tinyapp-controller:
	go build -mod vendor -o ${DIST_DIR}/tinyapp-controller -v ${CURRENT_DIR}/controller/cmd/main.go

tinyapp-controller-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 make tinyapp-controller

tinyapp-gateway:
	go build -mod vendor -o ${DIST_DIR}/tinyapp-gateway -v ${CURRENT_DIR}/gateway/cmd/main.go

tinyapp-gateway-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 make tinyapp-gateway
