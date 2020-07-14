// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package main

import (
	"context"

	"github.com/google/knative-gcp/pkg/broker/config/volume"
	"github.com/google/knative-gcp/pkg/broker/handler"
	"github.com/google/knative-gcp/pkg/metrics"
	"github.com/google/knative-gcp/pkg/utils/clients"
)

// Injectors from wire.go:

func InitializeSyncPool(ctx context.Context, projectID clients.ProjectID, podName metrics.PodName, containerName metrics.ContainerName, targetsVolumeOpts []volume.Option, opts ...handler.Option) (*handler.FanoutPool, error) {
	readonlyTargets, err := volume.NewTargetsFromFile(targetsVolumeOpts...)
	if err != nil {
		return nil, err
	}
	client, err := clients.NewPubsubClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	httpClient := _wireClientValue
	v := _wireValue
	retryClient, err := handler.NewRetryClient(ctx, client, v...)
	if err != nil {
		return nil, err
	}
	deliveryReporter, err := metrics.NewDeliveryReporter(podName, containerName)
	if err != nil {
		return nil, err
	}
	fanoutPool, err := handler.NewFanoutPool(readonlyTargets, client, httpClient, retryClient, deliveryReporter, opts...)
	if err != nil {
		return nil, err
	}
	return fanoutPool, nil
}

var (
	_wireClientValue = handler.DefaultHTTPClient
	_wireValue       = handler.DefaultCEClientOpts
)
