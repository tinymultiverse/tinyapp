#!/bin/bash

cd ./pkg/server/api/v1/proto

buf dep update
buf build
buf generate
