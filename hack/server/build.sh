#!/bin/bash

cd ./pkg/server/api/v1/proto

buf mod update
buf build
buf generate
