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

package v1

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	controllerutil "github.com/tinymultiverse/tinyapp/controller/util"
	"github.com/tinymultiverse/tinyapp/pkg/k8s/api/tinyapp/v1alpha1"
	pb "github.com/tinymultiverse/tinyapp/pkg/server/api/v1/proto"
	"github.com/tinymultiverse/tinyapp/server/util"
	globalutil "github.com/tinymultiverse/tinyapp/util"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	applycorev1 "k8s.io/client-go/applyconfigurations/core/v1"
	applymetav1 "k8s.io/client-go/applyconfigurations/meta/v1"
)

func (s *Server) CreateTinyApp(ctx context.Context, in *pb.CreateTinyAppRequest) (*pb.CreateTinyAppResponse, error) {
	if in == nil || in.AppDetail == nil {
		zap.S().Error("Input is empty for create request")
		return nil, errors.New("empty input")
	}

	// TODO more input validation

	logger := zap.S().With("appName", in.AppDetail.Name, "appType", in.AppDetail.AppType.String())
	logger.Info("Received request to create tiny app")

	tinyApp, err := s.deployTinyApp(ctx, in.AppDetail)
	if err != nil {
		logger.Errorw("Failed to create create tiny app", "error", err)
		return nil, err
	}

	appUrl, err := util.GetURLForTinyApp(s.env.AppIngressDomain, s.env.AppIngressSubPath, tinyApp.Name, s.env.AppIngressTlsEnabled)
	if err != nil {
		logger.Errorw("Failed to create app url", "error", err)
		return nil, err
	}

	logger.Infow("Successfully created tiny app", "App Id", tinyApp.Name)

	return &pb.CreateTinyAppResponse{
		AppRelease: &pb.TinyAppRelease{
			Id:                tinyApp.Name,
			AppUrl:            appUrl,
			CreationTimeStamp: tinyApp.CreationTimestamp.Time.String(),
			AppImage:          in.AppDetail.Image,
		},
	}, nil
}

func (s *Server) ListTinyApps(ctx context.Context, req *pb.ListTinyAppsRequest) (*pb.ListTinyAppsResponse, error) {
	logger := zap.S()
	logger.Info("Received request to list tiny apps")

	tinyApps, err := s.tinyAppClient.
		TinymultiverseV1alpha1().
		TinyApps(s.env.TinyAppNamespace).
		List(context.Background(), v1.ListOptions{})
	if err != nil {
		logger.Errorw("Failed to fetch TinyApps from K8s", "error", err)
		return nil, err
	}

	apps := make([]*pb.TinyApp, 0)

	for _, tinyApp := range tinyApps.Items {
		protoTinyApp, err := util.ConvertToProtoTinyApp(&tinyApp)
		if err != nil {
			logger.Errorw("Error while converting to proto TinyApp", "error", err)
			continue
		}

		apps = append(apps, protoTinyApp)
	}

	logger.Infof("Listed %d tiny apps", len(apps))

	return &pb.ListTinyAppsResponse{Apps: apps}, nil
}

func (s *Server) UpdateTinyApp(ctx context.Context, in *pb.UpdateTinyAppRequest) (*pb.UpdateTinyAppResponse, error) {
	logger := zap.S().With("appId", in.AppId)
	logger.Info("Received request to update tiny app")

	if in.AppId == "" {
		logger.Info("Empty app id")
		return nil, errors.New("empty app id")
	}

	if in.AppDetail == nil {
		logger.Info("Empty app detail")
		return nil, errors.New("empty app detail")
	}

	// First make sure app to update exists
	appsList, err := s.tinyAppClient.TinymultiverseV1alpha1().TinyApps(s.env.TinyAppNamespace).List(ctx, v1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", globalutil.K8sNameLabel, in.AppId),
	})
	if err != nil {
		logger.Errorw("Failed to get TinyApp", "error", err)
		return nil, err
	}
	if len(appsList.Items) == 0 {
		logger.Error("TinyApp not found")
		return nil, errors.New("TinyApp not found")
	}

	existingApp := appsList.Items[0]
	newApp, err := util.ConvertToK8sTinyApp(in.AppDetail, existingApp.Name, s.env)
	if err != nil {
		return nil, err
	}

	// Override existing app object's spec with new app spec
	existingApp.Spec = newApp.Spec

	if err := s.deploySecret(ctx, &existingApp.Name, in.AppDetail); err != nil {
		return nil, err
	}

	updatedApp, err := s.tinyAppClient.TinymultiverseV1alpha1().TinyApps(s.env.TinyAppNamespace).Update(ctx, &existingApp, v1.UpdateOptions{FieldManager: util.FieldManager})
	if err != nil {
		logger.Errorw("Failed to apply TinyApp updates to k8s", "error", err)
		return nil, err
	}

	appUrl, err := util.GetURLForTinyApp(s.env.AppIngressDomain, s.env.AppIngressSubPath, updatedApp.Name, s.env.AppIngressTlsEnabled)
	if err != nil {
		logger.Errorw("Failed to create app url", "error", err)
		return nil, err
	}

	logger.Infow("Successfully updated tiny app")

	return &pb.UpdateTinyAppResponse{
		AppRelease: &pb.TinyAppRelease{
			Id:                updatedApp.Name,
			AppUrl:            appUrl,
			CreationTimeStamp: updatedApp.CreationTimestamp.Time.String(),
			AppImage:          updatedApp.Spec.Image,
		},
	}, nil
}

func (s *Server) DeleteTinyApp(ctx context.Context, in *pb.DeleteTinyAppRequest) (*emptypb.Empty, error) {
	logger := zap.S().With("appId", in.AppId)
	logger.Info("Received request to delete TinyApp")

	// Execute deletion
	dp := v1.DeletePropagationBackground
	err := s.tinyAppClient.TinymultiverseV1alpha1().TinyApps(s.env.TinyAppNamespace).Delete(ctx, in.AppId,
		v1.DeleteOptions{PropagationPolicy: &dp})
	if err != nil {
		logger.Errorw("Failed to delete TinyApp", "error", err)
		return nil, err
	}

	err = s.k8sClient.CoreV1().Secrets(s.env.TinyAppNamespace).Delete(ctx, in.AppId, v1.DeleteOptions{})
	if err != nil && !k8sErrors.IsNotFound(err) {
		logger.Errorw("Failed to delete git token secret", "error", err)
	}

	zap.S().Infof("Successfully deleted tiny app")

	return &emptypb.Empty{}, nil
}

// deployTinyApp converts given TinyAppDetail into k8s TinyApp object and deploys it.
// Returns TinyApp k8s object if deployment is successful.
func (s *Server) deployTinyApp(ctx context.Context, appDetail *pb.TinyAppDetail) (*v1alpha1.TinyApp, error) {
	logger := zap.S().With("appName", appDetail.Name, "appType", appDetail.AppType.String())

	appObjName := globalutil.GenerateTinyAppObjName()
	newApp, err := util.ConvertToK8sTinyApp(appDetail, appObjName, s.env)
	if err != nil {
		return nil, err
	}

	if err := s.deploySecret(ctx, &newApp.Name, appDetail); err != nil {
		return nil, err
	}

	logger.Debug("creating TinyApp k8s object")
	tinyApp, err := s.tinyAppClient.TinymultiverseV1alpha1().TinyApps(s.env.TinyAppNamespace).Create(ctx, newApp, v1.CreateOptions{})
	if err != nil { // since previous lines deal with service but no tiny app, just fail
		return nil, err
	}

	return tinyApp, nil
}

// deploySecret deploys k8s secret containing git token if applicable.
func (s *Server) deploySecret(ctx context.Context, name *string, appDetail *pb.TinyAppDetail) error {
	kind := "Secret"
	apiVersion := "v1"
	if appDetail.SourceType == pb.SourceType_SOURCE_TYPE_GIT && appDetail.GitConfig.Token != "" {
		secretApplyConfig := &applycorev1.SecretApplyConfiguration{
			TypeMetaApplyConfiguration: applymetav1.TypeMetaApplyConfiguration{
				Kind:       &kind,
				APIVersion: &apiVersion,
			},
			ObjectMetaApplyConfiguration: &applymetav1.ObjectMetaApplyConfiguration{
				Name:      name,
				Namespace: &s.env.TinyAppNamespace,
			},
			StringData: map[string]string{
				controllerutil.GitTokenSecretKey: appDetail.GitConfig.Token,
			},
		}

		_, err := s.k8sClient.CoreV1().Secrets(s.env.TinyAppNamespace).Apply(ctx, secretApplyConfig, v1.ApplyOptions{FieldManager: util.FieldManager})
		if err != nil {
			return errors.WithMessage(err, "failed to apply git token secret")
		}
	}

	return nil
}

func (s *Server) GetTinyApp(ctx context.Context, appId string) (*v1alpha1.TinyApp, error) {
	tinyApp, err := s.tinyAppClient.TinymultiverseV1alpha1().TinyApps(s.env.TinyAppNamespace).Get(ctx, appId, v1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to load TinyApp %s", appId))
	}
	return tinyApp, nil
}
