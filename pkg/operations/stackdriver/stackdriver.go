/*
Copyright 2019 Google LLC.

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

package operations

import (
	"context"
	"fmt"

	"cloud.google.com/go/compute/metadata"
	"cloud.google.com/go/logging/logadmin"
	corev1 "k8s.io/api/core/v1"
)

type StackdriverArgs struct {
	// ProjectId in which to configure the Stackdriver sink. Other
	// parent resources are not yet supported.
	ProjectID string
}

func (_ StackdriverArgs) OperationGroup() string {
	return "stackdriver"
}

type StackdriverOps struct {
	// Environment variable containing project id.
	Project string `envconfig:"PROJECT_ID"`

	// Client is the Stackdriver client used by Ops.
	Client *logadmin.Client
}

func StackdriverEnv(args StackdriverArgs) []corev1.EnvVar {
	return []corev1.EnvVar{{
		Name:  "PROJECT_ID",
		Value: args.ProjectID,
	}}
}

func ValidateStackdriverArgs(args StackdriverArgs) error {
	if args.ProjectID == "" {
		return fmt.Errorf("missing ProjectID")
	}
	return nil
}

func (s *StackdriverOps) CreateClient(ctx context.Context) error {
	if s.Project == "" {
		project, err := metadata.ProjectID()
		if err != nil {
			return err
		}
		s.Project = project
	}

	// TODO: support other parent resources.
	client, err := logadmin.NewClient(ctx, s.Project)
	if err != nil {
		return fmt.Errorf("failed to create Stackdriver client: %s", err)
	}
	s.Client = client
	return nil
}
