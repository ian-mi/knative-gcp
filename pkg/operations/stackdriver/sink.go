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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"cloud.google.com/go/logging/logadmin"
	"cloud.google.com/go/pubsub"
	"github.com/google/knative-gcp/pkg/operations"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"knative.dev/pkg/logging"

	corev1 "k8s.io/api/core/v1"
)

type SinkActionResult struct {
	Result bool   `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`

	// Project is the project id that we used (this might have
	// been defaulted, to we'll expose it).
	ProjectId string `json:"projectId,omitempty"`
}

type SinkArgs struct {
	StackdriverArgs
	// Id of the sink to act upon.
	SinkID string
}

func (_ SinkArgs) LabelKey() string {
	return "sink"
}

func (_ SinkArgs) OperationSubgroup() string {
	return "sink"
}

func SinkEnv(args SinkArgs) []corev1.EnvVar {
	return append(StackdriverEnv(args.StackdriverArgs),
		corev1.EnvVar{
			Name:  "SINK_ID",
			Value: args.SinkID,
		},
	)

}

func ValidateSinkArgs(args SinkArgs) error {
	if args.SinkID == "" {
		return fmt.Errorf("missing SinkId")
	}
	return ValidateStackdriverArgs(args.StackdriverArgs)
}

type SinkCreateArgs struct {
	SinkArgs
	// TopicId to use as sink destination. Must belong to the
	// ProjectID specified above.
	TopicID string
	// Filter is an optional loggin query filter.
	Filter string
}

func (_ SinkCreateArgs) Action() string {
	return operations.ActionCreate
}

func (a SinkCreateArgs) Env() []corev1.EnvVar {
	return append(SinkEnv(a.SinkArgs),
		corev1.EnvVar{
			Name:  "STACKDRIVER_FILTER",
			Value: a.Filter,
		},
		corev1.EnvVar{
			Name:  "PUBSUB_TOPIC_ID",
			Value: a.TopicID,
		},
	)

}

func (a SinkCreateArgs) Validate() error {
	if a.TopicID == "" {
		return fmt.Errorf("missing TopicID")
	}
	return ValidateSinkArgs(a.SinkArgs)
}

type SinkDeleteArgs struct {
	SinkArgs
}

func (_ SinkDeleteArgs) Action() string {
	return operations.ActionDelete
}

func (a SinkDeleteArgs) Env() []corev1.EnvVar {
	return SinkEnv(a.SinkArgs)
}

func (a SinkDeleteArgs) Validate() error {
	return ValidateSinkArgs(a.SinkArgs)
}

type SinkOps struct {
	StackdriverOps

	// Sink action to run.
	// Options: [create]
	Action string `envconfig:"ACTION" required:"true"`

	// ID of the sink to act upon.
	SinkId string `envconfig:"SINK_ID" required:"true"`

	// Topic is the environment variable containing the PubSub
	// Topic being used as the sink's destination. In the form that is
	// unique within the project.  E.g. 'laconia', not
	// 'projects/my-gcp-project/topics/laconia'.
	Topic string `envconfig:"PUBSUB_TOPIC_ID" required:"false"`

	// Filter is an optional logging query to use with the sink.
	Filter string `envconfig:"STACKDRIVER_FILTER" required:"false" default:""`
}

func (s *SinkOps) Run(ctx context.Context) error {
	if s.Client == nil {
		return errors.New("Stackdriver client is nil")
	}
	logger := logging.FromContext(ctx)

	logger = logger.With(
		zap.String("action", s.Action),
		zap.String("sink", s.SinkId),
		zap.String("project", s.Project),
	)
	logger.Info("Stackdriver Sink Job.")

	result := SinkActionResult{
		ProjectId: s.Project,
	}
	defer func() {
		if err := writeTerminationMessage(&result); err != nil {
			logger.Infof("Failed to write termination message: %s", err)
		}
	}()

	switch s.Action {
	case operations.ActionCreate:
		// Create a pubsub client in order to grant publisher
		// role to the sink writer identity.
		pubsubClient, err := pubsub.NewClient(ctx, s.Project)
		if err != nil {
			result.Result = false
			result.Error = err.Error()
			return err
		}

		sink := &logadmin.Sink{
			ID:          s.SinkId,
			Destination: fmt.Sprintf("pubsub.googleapis.com/projects/%s/topics/%s", s.Project, s.Topic),
			Filter:      s.Filter,
		}
		sink, err = s.Client.CreateSinkOpt(ctx, sink, logadmin.SinkOptions{UniqueWriterIdentity: true})
		// TODO: handle sink already exists.
		if err != nil {
			result.Result = false
			result.Error = err.Error()
			return err
		}
		logger.Infof("Created Sink %s.", sink.ID)
		// Update topic IAM using the sink's WriteIdentity.
		// TODO: handle by topic reconciler in separate operation.
		topicIam := pubsubClient.Topic(s.Topic).IAM()
		topicPolicy, err := topicIam.Policy(ctx)
		if err != nil {
			result.Result = false
			result.Error = err.Error()
			return err
		}
		topicPolicy.Add(sink.WriterIdentity, "roles/pubsub.publisher")
		if err = topicIam.SetPolicy(ctx, topicPolicy); err != nil {
			result.Result = false
			result.Error = err.Error()
			return err
		}
		logger.Infof("Added gave writer identify '%s' roles/pubsub.publisher on topic '%s'.", sink.WriterIdentity, s.Topic)
		result.Result = true
	case operations.ActionDelete:
		if err := s.Client.DeleteSink(ctx, s.SinkId); err != nil {
			if st, isStatus := status.FromError(err); isStatus && st.Code() == codes.NotFound {
				logger.Info("Sink %s not found.", s.SinkId)
			} else {
				result.Result = false
				result.Error = err.Error()
			}
		} else {
			logger.Info("Sink %s deleted.", s.SinkId)
		}
		result.Result = true
	default:
		return fmt.Errorf("Unknown action value %v", s.Action)
	}

	logger.Info("Done.")
	return nil
}

func writeTerminationMessage(result *SinkActionResult) error {
	m, err := json.Marshal(result)
	if err != nil {
		return err
	}
	return ioutil.WriteFile("/dev/termination-log", m, 0644)
}
