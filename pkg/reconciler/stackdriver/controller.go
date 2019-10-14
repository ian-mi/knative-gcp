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

package stackdriver

import (
	"context"

	"github.com/google/knative-gcp/pkg/apis/duck/v1alpha1"
	"github.com/google/knative-gcp/pkg/pubsub/adapter/converters"
	"github.com/google/knative-gcp/pkg/reconciler"
	"github.com/google/knative-gcp/pkg/reconciler/job"
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
	"k8s.io/client-go/tools/cache"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"

	stackdriverinformers "github.com/google/knative-gcp/pkg/client/injection/informers/events/v1alpha1/stackdriver"
	pullsubscriptioninformers "github.com/google/knative-gcp/pkg/client/injection/informers/pubsub/v1alpha1/pullsubscription"
	topicinformers "github.com/google/knative-gcp/pkg/client/injection/informers/pubsub/v1alpha1/topic"
	jobinformer "knative.dev/pkg/client/injection/kube/informers/batch/v1/job"
)

const (
	// controllerAgentName is the string used by this controller to identify
	// itself when creating events.
	controllerAgentName = "cloud-run-events-stackdriver-source-controller"
)

type envConfig struct {
	// SinkOpsImage is the image for operating on stackdriver sinks. Required.
	SinkOpsImage string `envconfig:"STACKDRIVER_SINK_IMAGE" required:"true"`
}

// NewController initializes the controller and is called by the generated code
// Registers event handlers to enqueue events
func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {

	pullsubscriptionInformer := pullsubscriptioninformers.Get(ctx)
	topicInformer := topicinformers.Get(ctx)
	jobInformer := jobinformer.Get(ctx)
	stackdriverInformer := stackdriverinformers.Get(ctx)

	logger := logging.FromContext(ctx).Named(controllerAgentName)
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		logger.Fatal("Failed to process env var", zap.Error(err))
	}

	c := &Reconciler{
		SinkOpsImage:      env.SinkOpsImage,
		PubSubBase:        reconciler.NewPubSubBase(ctx, controllerAgentName, "stackdriver.events.cloud.google.com", converters.StackdriverAdapterType, cmw),
		stackdriverLister: stackdriverInformer.Lister(),
	}
	c.jobReconciler = job.Reconciler{
		KubeClientSet: c.KubeClientSet,
		JobLister:     jobInformer.Lister(),
		Logger:        c.Logger,
	}
	impl := controller.NewImpl(c, c.Logger, ReconcilerName)

	c.Logger.Info("Setting up event handlers")
	stackdriverInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	jobInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.Filter(v1alpha1.SchemeGroupVersion.WithKind("Stackdriver")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	topicInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.Filter(v1alpha1.SchemeGroupVersion.WithKind("Stackdriver")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	pullsubscriptionInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.Filter(v1alpha1.SchemeGroupVersion.WithKind("Stackdriver")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	//	c.tracker = tracker.New(impl.EnqueueKey, controller.GetTrackerLease(ctx))

	return impl
}
