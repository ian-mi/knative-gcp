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
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/knative-gcp/pkg/apis/events/v1alpha1"
	listers "github.com/google/knative-gcp/pkg/client/listers/events/v1alpha1"
	ops "github.com/google/knative-gcp/pkg/operations"
	operations "github.com/google/knative-gcp/pkg/operations/stackdriver"
	"github.com/google/knative-gcp/pkg/reconciler"
	"github.com/google/knative-gcp/pkg/reconciler/job"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/cache"
	"knative.dev/pkg/logging"
)

const (
	ReconcilerName = "Stackdriver"

	finalizerName = controllerAgentName

	resourceGroup = "stackdrivers.events.cloud.google.com"
)

type Reconciler struct {
	*reconciler.PubSubBase

	// Image to use for stackdriver sink operations.
	SinkOpsImage string

	stackdriverLister listers.StackdriverLister

	jobReconciler job.Reconciler
}

func (c *Reconciler) Reconcile(ctx context.Context, key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the Storage resource with this namespace/name
	original, err := c.stackdriverLister.Stackdrivers(namespace).Get(name)
	if apierrs.IsNotFound(err) {
		// The Storage resource may no longer exist, in which case we stop processing.
		runtime.HandleError(fmt.Errorf("stackdriver '%s' in work queue no longer exists", key))
		return nil
	} else if err != nil {
		return err
	}

	// Don't modify the informers copy
	csr := original.DeepCopy()

	reconcileErr := c.reconcile(ctx, csr)

	if equality.Semantic.DeepEqual(original.Status, csr.Status) &&
		equality.Semantic.DeepEqual(original.ObjectMeta, csr.ObjectMeta) {
		// If we didn't change anything then don't call updateStatus.
		// This is important because the copy we loaded from the informer's
		// cache may be stale and we don't want to overwrite a prior update
		// to status with this stale state.
	} else if _, err := c.updateStatus(ctx, csr); err != nil {
		// TODO: record the event (c.Recorder.Eventf(...
		c.Logger.Warn("Failed to update Stackdriver Source status", zap.Error(err))
		return err
	}

	if reconcileErr != nil {
		// TODO: record the event (c.Recorder.Eventf(...
		return reconcileErr
	}

	return nil
}

func (c *Reconciler) reconcile(ctx context.Context, csr *v1alpha1.Stackdriver) error {
	csr.Status.ObservedGeneration = csr.Generation
	topic := csr.Status.TopicID
	if topic == "" {
		topic = fmt.Sprintf("stackdriver-%s", string(csr.UID))
	}
	sinkID := csr.Status.SinkID
	if sinkID == "" {
		sinkID = fmt.Sprintf("stackdriver-%s", string(csr.UID))
	}

	csr.Status.InitializeConditions()

	// See if the source has been deleted.
	deletionTimestamp := csr.DeletionTimestamp

	if deletionTimestamp != nil {
		err := c.deleteSink(ctx, csr)
		if err != nil {
			c.Logger.Infof("Unable to delete the Sink: %s", err)
			return err
		}
		err = c.PubSubBase.DeletePubSub(ctx, csr.Namespace, csr.Name)
		if err != nil {
			c.Logger.Infof("Unable to delete pubsub resources : %s", err)
			return fmt.Errorf("failed to delete pubsub resources: %s", err)
		}
		c.removeFinalizer(csr)
		return nil
	}

	// Ensure that there's finalizer there, since we're about to attempt to
	// change external state with the topic, so we need to clean it up.
	err := c.ensureFinalizer(csr)
	if err != nil {
		return err
	}

	t, ps, err := c.PubSubBase.ReconcilePubSub(ctx, csr, topic, resourceGroup)
	if err != nil {
		c.Logger.Infof("Failed to reconcile PubSub: %s", err)
		return err
	}

	c.Logger.Infof("Reconciled: PubSub: %+v PullSubscription: %+v", t, ps)

	sink, err := c.reconcileSink(ctx, csr)
	if err != nil {
		// TODO: Update status with this...
		c.Logger.Infof("Failed to reconcile Stackdriver Sink: %s", err)
		csr.Status.MarkSinkNotReady("SinkNotReady", "Failed to create Stackdriver sink: %s", err)
		return err
	}

	csr.Status.MarkSinkReady()

	c.Logger.Infof("Reconciled Stackdriver sink: %+v", sink)
	csr.Status.SinkID = sink
	return nil
}

func (c *Reconciler) reconcileSink(ctx context.Context, stackdriver *v1alpha1.Stackdriver) (string, error) {
	sinkID := fmt.Sprintf("sink-%s", string(stackdriver.UID))
	createArgs := operations.SinkCreateArgs{
		SinkArgs: operations.SinkArgs{
			StackdriverArgs: operations.StackdriverArgs{
				ProjectID: stackdriver.Status.ProjectID,
			},
			SinkID: sinkID,
		},
		TopicID: stackdriver.Status.TopicID,
		Filter:  stackdriver.Spec.Filter,
	}
	state, err := c.jobReconciler.EnsureOpJob(ctx, c.sinkOpCtx(stackdriver), createArgs)
	if err != nil {
		return "", fmt.Errorf("Error creating job %q: %q", stackdriver.Name, err)
	}

	if state == ops.OpsJobCreateFailed || state == ops.OpsJobCompleteFailed {
		return "", fmt.Errorf("Job %q failed to create or job failed", stackdriver.Name)
	}

	if state != ops.OpsJobCompleteSuccessful {
		return "", fmt.Errorf("Job %q has not completed yet", stackdriver.Name)
	}

	// See if the pod exists or not...
	pod, err := ops.GetJobPod(ctx, c.KubeClientSet, stackdriver.Namespace, string(stackdriver.UID), ops.ActionCreate)
	if err != nil {
		return "", err
	}

	var result operations.SinkActionResult
	if err := ops.GetOperationsResult(ctx, pod, &result); err != nil {
		return "", err
	}
	if result.Result {
		return sinkID, nil
	}
	return "", fmt.Errorf("operation failed: %s", result.Error)
}

// deleteSink looks at status.SinkID and if non-empty will delete the
// previously created stackdriver sink.
func (c *Reconciler) deleteSink(ctx context.Context, stackdriver *v1alpha1.Stackdriver) error {
	if stackdriver.Status.SinkID == "" {
		return nil
	}

	deleteArgs := operations.SinkDeleteArgs{
		SinkArgs: operations.SinkArgs{
			StackdriverArgs: operations.StackdriverArgs{
				ProjectID: stackdriver.Status.ProjectID,
			},
			SinkID: stackdriver.Status.SinkID,
		},
	}
	state, err := c.jobReconciler.EnsureOpJob(ctx, c.sinkOpCtx(stackdriver), deleteArgs)
	if err != nil {
		return fmt.Errorf("Error creating job %q: %q", stackdriver.Name, err)
	}

	if state != ops.OpsJobCompleteSuccessful {
		return fmt.Errorf("Job %q has not completed yet", stackdriver.Name)
	}

	// See if the pod exists or not...
	pod, err := ops.GetJobPod(ctx, c.KubeClientSet, stackdriver.Namespace, string(stackdriver.UID), ops.ActionDelete)
	if err != nil {
		return err
	}
	// Check to see if the operation worked
	var result operations.SinkActionResult
	if err := ops.GetOperationsResult(ctx, pod, &result); err != nil {
		return err
	}
	if !result.Result {
		return fmt.Errorf("operation failed: %s", result.Error)
	}

	c.Logger.Infof("Deleted Sink: %q", stackdriver.Status.SinkID)
	stackdriver.Status.SinkID = ""
	return nil
}

func (c *Reconciler) ensureFinalizer(csr *v1alpha1.Stackdriver) error {
	finalizers := sets.NewString(csr.Finalizers...)
	if finalizers.Has(finalizerName) {
		return nil
	}
	mergePatch := map[string]interface{}{
		"metadata": map[string]interface{}{
			"finalizers":      append(csr.Finalizers, finalizerName),
			"resourceVersion": csr.ResourceVersion,
		},
	}
	patch, err := json.Marshal(mergePatch)
	if err != nil {
		return err
	}
	_, err = c.RunClientSet.EventsV1alpha1().Stackdrivers(csr.Namespace).Patch(csr.Name, types.MergePatchType, patch)
	return err

}

func (c *Reconciler) removeFinalizer(csr *v1alpha1.Stackdriver) error {
	// Only remove our finalizer if it's the first one.
	if len(csr.Finalizers) == 0 || csr.Finalizers[0] != finalizerName {
		return nil
	}

	// For parity with merge patch for adding, also use patch for removing
	mergePatch := map[string]interface{}{
		"metadata": map[string]interface{}{
			"finalizers":      csr.Finalizers[1:],
			"resourceVersion": csr.ResourceVersion,
		},
	}
	patch, err := json.Marshal(mergePatch)
	if err != nil {
		return err
	}
	_, err = c.RunClientSet.EventsV1alpha1().Stackdrivers(csr.Namespace).Patch(csr.Name, types.MergePatchType, patch)
	return err
}

func (c *Reconciler) updateStatus(ctx context.Context, desired *v1alpha1.Stackdriver) (*v1alpha1.Stackdriver, error) {
	source, err := c.stackdriverLister.Stackdrivers(desired.Namespace).Get(desired.Name)
	if err != nil {
		return nil, err
	}
	// Check if there is anything to update.
	if equality.Semantic.DeepEqual(source.Status, desired.Status) {
		return source, nil
	}
	becomesReady := desired.Status.IsReady() && !source.Status.IsReady()

	// Don't modify the informers copy.
	existing := source.DeepCopy()
	existing.Status = desired.Status
	src, err := c.RunClientSet.EventsV1alpha1().Stackdrivers(desired.Namespace).UpdateStatus(existing)

	if err == nil && becomesReady {
		duration := time.Since(src.ObjectMeta.CreationTimestamp.Time)
		c.Logger.Infof("Stackdriver %q became ready after %v", source.Name, duration)

		if err := c.StatsReporter.ReportReady("Stackdriver", source.Namespace, source.Name, duration); err != nil {
			logging.FromContext(ctx).Infof("failed to record ready for Stackdriver, %v", err)
		}
	}

	return src, err
}

func (c *Reconciler) sinkOpCtx(stackdriver *v1alpha1.Stackdriver) ops.OpCtx {
	return ops.OpCtx{
		Image:  c.SinkOpsImage,
		UID:    string(stackdriver.UID),
		Secret: *stackdriver.Spec.Secret,
		Owner:  stackdriver,
	}
}
