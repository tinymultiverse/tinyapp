FROM golang:1.22.4 AS builder

RUN apt-get update && apt-get install -y ca-certificates && update-ca-certificates

RUN mkdir /tinyapp
COPY . /tinyapp
WORKDIR /tinyapp

RUN go mod vendor
RUN make tinyapp-linux && chmod +x dist/tinyapp

FROM golang:1.22.4
COPY --from=builder /tinyapp/dist/tinyapp /bin/tinyapp
ENTRYPOINT ["/bin/tinyapp"]
