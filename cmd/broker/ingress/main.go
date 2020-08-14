/*
Copyright 2020 Google LLC

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
	"errors"
	"time"

	metadataClient "github.com/google/knative-gcp/pkg/gclient/metadata"
	"github.com/google/knative-gcp/pkg/metrics"
	"github.com/google/knative-gcp/pkg/utils"
	"github.com/google/knative-gcp/pkg/utils/appcredentials"
	"github.com/google/knative-gcp/pkg/utils/clients"
	"github.com/google/knative-gcp/pkg/utils/mainhelper"
	"knative.dev/eventing/pkg/kncloudevents"

	"go.uber.org/zap"
)

type envConfig struct {
	PodName   string `envconfig:"POD_NAME" required:"true"`
	Port      int    `envconfig:"PORT" default:"8080"`
	ProjectID string `envconfig:"PROJECT_ID"`
}

const (
	component       = "broker-ingress"
	metricNamespace = "broker"
)

// main creates and starts an ingress handler using default options.
// 1. It listens on port specified by "PORT" env var, or default 8080 if env var is not set
// 2. It reads "PROJECT_ID" env var for pubsub project. If the env var is empty, it retrieves project ID from
//    GCE metadata.
// 3. It expects broker configmap mounted at "/var/run/cloud-run-events/broker/targets"
func main() {
	appcredentials.MustExistOrUnsetEnv()

	var env envConfig
	ctx, res := mainhelper.Init(component, mainhelper.WithMetricNamespace(metricNamespace), mainhelper.WithEnv(&env))
	defer res.Cleanup()
	logger := res.Logger

	projectID, err := utils.ProjectID(env.ProjectID, metadataClient.NewDefaultMetadataClient())
	if err != nil {
		logger.Desugar().Fatal("Failed to create project id", zap.Error(err))
	}
	logger.Desugar().Info("Starting ingress handler", zap.Any("envConfig", env), zap.Any("Project ID", projectID))

	ingress, err := InitializeHandler(
		ctx,
		clients.Port(env.Port),
		clients.ProjectID(projectID),
		metrics.PodName(env.PodName),
		metrics.ContainerName(component),
	)
	if err != nil {
		logger.Desugar().Fatal("Unable to create ingress handler: ", zap.Error(err))
	}

	logger.Desugar().Info("Starting ingress.", zap.Any("ingress", ingress))
	if err := ingress.Start(kncloudevents.WithShutdownTimeout(ctx, 15*time.Second)); err != nil {
		if errors.Is(err, context.Canceled) {
			logger.Desugar().Info("Stopped ingress.")
		} else {
			logger.Desugar().Error("Ingress stopped with error.", zap.Error(err))
		}
	}
}
