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
	"flag"
	"log"
	"os"
	"time"

	"cloud.google.com/go/pubsub"
	"go.uber.org/zap"

	"knative.dev/pkg/metrics"

	"github.com/google/knative-gcp/pkg/broker/config/volume"
	"github.com/google/knative-gcp/pkg/broker/handler/pool"
	metadataClient "github.com/google/knative-gcp/pkg/gclient/metadata"
	"github.com/google/knative-gcp/pkg/utils"
	"github.com/google/knative-gcp/pkg/utils/appcredentials"
	"github.com/google/knative-gcp/pkg/utils/mainhelper"
)

const (
	component = "broker-retry"
)

type envConfig struct {
	PodName            string `envconfig:"POD_NAME" required:"true"`
	ProjectID          string `envconfig:"PROJECT_ID"`
	TargetsConfigPath  string `envconfig:"TARGETS_CONFIG_PATH" default:"/var/run/cloud-run-events/broker/targets"`
	HandlerConcurrency int    `envconfig:"HANDLER_CONCURRENCY"`

	// Outstanding messages effectively limits how many connections we will create to each subscriber.
	// If such connections are long, it will consume a lot of memory (aggregated) without limiting.
	OutstandingMessagesPerSub int `envconfig:"OUTSTANDING_MESSAGES_PER_SUB" default:"100"`
	// 3Mi. We also want to limit the memory usage from each subscription.
	OutstandingBytesPerSub int `envconfig:"OUTSTANDING_BYTES_PER_SUB" default:"3000000"`

	// Max to 10m.
	TimeoutPerEvent time.Duration `envconfig:"TIMEOUT_PER_EVENT"`
}

func main() {
	appcredentials.MustExistOrUnsetEnv()
	flag.Parse()

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}
	var env envConfig
	ctx, res := mainhelper.Init(
		component,
		mainhelper.WithEnv(&env),
		// There is no retry service so use the pod's hostname
		mainhelper.WithServiceName(hostname),
	)
	defer res.Cleanup()
	logger := res.Logger

	// Give the signal channel some buffer so that reconciling handlers won't
	// block the targets config update?
	targetsUpdateCh := make(chan struct{})
	logger.Info("Starting the broker retry")

	projectID, err := utils.ProjectID(env.ProjectID, metadataClient.NewDefaultMetadataClient())
	if err != nil {
		logger.Fatalf("failed to get default ProjectID: %v", err)
	}

	syncSignal := poolSyncSignal(ctx, targetsUpdateCh)
	syncPool, err := InitializeSyncPool(
		ctx,
		pool.ProjectID(projectID),
		[]volume.Option{
			volume.WithPath(env.TargetsConfigPath),
			volume.WithNotifyChan(targetsUpdateCh),
		},
		buildPoolOptions(env)...,
	)
	if err != nil {
		logger.Fatal("Failed to get retry sync pool", zap.Error(err))
	}
	if _, err := pool.StartSyncPool(ctx, syncPool, syncSignal); err != nil {
		logger.Fatal("Failed to start retry sync pool", zap.Error(err))
	}

	// Context will be done if a TERM signal is issued.
	<-ctx.Done()
	logger.Info("Exiting...")
}

func poolSyncSignal(ctx context.Context, targetsUpdateCh chan struct{}) chan struct{} {
	// Give it some buffer so that multiple signal could queue up
	// but not blocking the signaler?
	ch := make(chan struct{}, 10)
	ticker := time.NewTicker(15 * time.Second)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-targetsUpdateCh:
				ch <- struct{}{}
			case <-ticker.C:
				ch <- struct{}{}
			}
		}
	}()
	return ch
}

func buildPoolOptions(env envConfig) []pool.Option {
	rs := pubsub.DefaultReceiveSettings
	// If Synchronous is true, then no more than MaxOutstandingMessages will be in memory at one time.
	// MaxOutstandingBytes still refers to the total bytes processed, rather than in memory.
	// NumGoroutines is ignored.
	// TODO Need to revisit it. For the case when synchronous is true, default value of MaxOutstandingMessages and MaxOutstandingBytes might need to override.
	rs.Synchronous = true
	rs.MaxOutstandingMessages = env.OutstandingMessagesPerSub
	rs.MaxOutstandingBytes = env.OutstandingBytesPerSub
	var opts []pool.Option
	if env.HandlerConcurrency > 0 {
		opts = append(opts, pool.WithHandlerConcurrency(env.HandlerConcurrency))
		rs.NumGoroutines = env.HandlerConcurrency
	}
	if env.TimeoutPerEvent > 0 {
		opts = append(opts, pool.WithTimeoutPerEvent(env.TimeoutPerEvent))
	}
	opts = append(opts, pool.WithPubsubReceiveSettings(rs))
	// The default CeClient is good?
	return opts
}

func flush(logger *zap.SugaredLogger) {
	_ = logger.Sync()
	metrics.FlushExporter()
}
