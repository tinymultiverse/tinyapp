FROM golang:1.22.4 as builder

RUN mkdir /tinyapp
COPY . /tinyapp
WORKDIR /tinyapp

RUN go mod vendor
RUN make tinyapp-linux && chmod +x dist/tinyapp

FROM golang:1.22.4
COPY --from=builder /tinyapp/dist/tinyapp-server /bin/tinyapp
ENTRYPOINT ["/bin/bash", "-c", "tinyapp"]
