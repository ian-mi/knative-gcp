// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package main

import (
	"context"

	"github.com/google/knative-gcp/pkg/apis/configs/gcpauth"
	"knative.dev/pkg/injection"
)

// Injectors from wire.go:

func InitializeControllers(ctx context.Context) ([]injection.ControllerConstructor, error) {
	storeSingleton := &gcpauth.StoreSingleton{}
	mainConversionController := newConversionConstructor(storeSingleton)
	mainDefaultingAdmissionController := newDefaultingAdmissionConstructor(storeSingleton)
	mainValidationController := newValidationConstructor(storeSingleton)
	v := Controllers(mainConversionController, mainDefaultingAdmissionController, mainValidationController)
	return v, nil
}
