/*
Copyright 2024 BlackRock, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"fmt"
	"net"
	"net/http"

	proto2 "github.com/tinymultiverse/tinyapp/pkg/server/api/v1/proto"
	"github.com/tinymultiverse/tinyapp/server/internal"
	"github.com/tinymultiverse/tinyapp/server/v1"
	"github.com/tinymultiverse/tinyapp/util/logging"

	"github.com/caarlos0/env/v10"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc" // Fix 'no Auth Provider found for name \"oidc\"'
)

func main() {
	logging.InitLoggerFromEnvironment()

	// Initialize environment variables
	envVars := internal.EnvVars{}
	if err := env.Parse(&envVars); err != nil {
		zap.S().Fatalf("could not process environment variables: %v", err)
	}

	zap.S().Info("tinyapp-server starting")
	lis, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", envVars.GRPCPort))
	if err != nil {
		zap.S().Fatalf("failed to listen: %v", err)
	}

	server, err := v1.NewServer(envVars)
	if err != nil {
		zap.S().Fatalf("failed to create server: %v", err)
	}

	s := grpc.NewServer()
	proto2.RegisterTinyAppServerServer(s, server)

	// Start up gRPC and REST servers
	go func() {
		zap.S().Infof("starting gRPC server on port %d", envVars.GRPCPort)
		if err := s.Serve(lis); err != nil {
			zap.S().Fatalf("failed to serve: %v", err)
		}
	}()

	createAndRunHttpServer(envVars)
}

func createAndRunHttpServer(envVars internal.EnvVars) {
	mux := runtime.NewServeMux()
	ctx := context.Background()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := proto2.RegisterTinyAppServerHandlerFromEndpoint(ctx, mux, fmt.Sprintf("127.0.0.1:%d", envVars.GRPCPort), opts)
	if err != nil {
		zap.S().Fatal(err)
	}

	zap.S().Infof("starting http server on port %d", envVars.HTTPPort)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", envVars.HTTPPort), mux); err != nil {
		zap.S().Fatal(err)
	}
}
