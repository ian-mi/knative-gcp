/*
Copyright 2019 Google LLC

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

package converters

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"strings"

	cloudevents "github.com/cloudevents/sdk-go"
	cepubsub "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/pubsub"
	pubsubcontext "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/pubsub/context"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/empty"
	"go.uber.org/zap"
	auditpb "google.golang.org/genproto/googleapis/cloud/audit"
	logpb "google.golang.org/genproto/googleapis/logging/v2"
	"knative.dev/pkg/logging"
)

const (
	StackdriverAdapterType = "google.stackdriver"

	logEntrySchema   = "type.googleapis.com/google.logging.v2.LogEntry"
	loggingSource    = "logging.googleapis.com"
	cloudAuditSource = "cloudaudit.googleapis.com"
)

var (
	logentryUnmarshaller = jsonpb.Unmarshaler{
		AllowUnknownFields: true,
		AnyResolver:        resolver(resolveAnyUnknowns),
	}
)

type resolver func(turl string) (proto.Message, error)

func (r resolver) Resolve(turl string) (proto.Message, error) {
	return r(turl)
}

type UnknownMsg empty.Empty

func (m *UnknownMsg) ProtoMessage() {
	(*empty.Empty)(m).ProtoMessage()
}

func (m *UnknownMsg) Reset() {
	(*empty.Empty)(m).Reset()
}

func (m *UnknownMsg) String() string {
	return "Unknown message"
}

func resolveAnyUnknowns(typeUrl string) (proto.Message, error) {
	// Only the part of typeUrl after the last slash is relevant.
	mname := typeUrl
	if slash := strings.LastIndex(mname, "/"); slash >= 0 {
		mname = mname[slash+1:]
	}
	mt := proto.MessageType(mname)
	if mt == nil {
		return (*UnknownMsg)(&empty.Empty{}), nil
	}
	return reflect.New(mt.Elem()).Interface().(proto.Message), nil
}

func convertStackdriver(ctx context.Context, msg *cepubsub.Message, sendMode ModeType) (*cloudevents.Event, error) {
	logger := logging.FromContext(ctx)
	if msg == nil {
		return nil, fmt.Errorf("nil pubsub message")
	}
	entry := logpb.LogEntry{}
	if err := logentryUnmarshaller.Unmarshal(bytes.NewReader(msg.Data), &entry); err != nil {
		return nil, fmt.Errorf("Failed to decode LogEntry: %q", err)
	}
	logger = logger.With(zap.Any("LogEntry.LogName", entry.LogName), zap.Any("LogEntry.Resource", entry.Resource))
	logger.Info("Processing Stackdriver LogEntry.")
	tx := pubsubcontext.TransportContextFrom(ctx)
	// Make a new event and convert the message payload.
	event := cloudevents.NewEvent(cloudevents.VersionV03)
	//TODO: Derive ID from entry
	event.SetID(tx.ID)
	if timestamp, err := ptypes.Timestamp(entry.Timestamp); err != nil {
		return nil, fmt.Errorf("Invalid LogEntry timestamp: %q", err)
	} else {
		event.SetTime(timestamp)
	}
	event.SetSchemaURL(logEntrySchema)
	event.SetDataContentType(cloudevents.ApplicationJSON)
	event.SetData(msg.Data)
	switch payload := entry.Payload.(type) {
	case *logpb.LogEntry_ProtoPayload:
		var unpacked ptypes.DynamicAny
		if err := ptypes.UnmarshalAny(payload.ProtoPayload, &unpacked); err != nil {
			return nil, fmt.Errorf("Unrecognized proto payload: %q:", err)
		}
		// TODO: handle AppEngine proto payloads.
		switch proto := unpacked.Message.(type) {
		case *auditpb.AuditLog:
			logger = logger.With(
				zap.Any("AuditLog.ServiceName", proto.ServiceName),
				zap.Any("AuditLog.ResourceName", proto.ResourceName),
				zap.Any("AuditLog.MethodName", proto.MethodName))
			logger.Info("Processing AuditLog.")
			event.SetSource(cloudAuditSource)
			event.SetSubject(proto.ServiceName + "/" + proto.ResourceName)
			event.SetType("com." + proto.MethodName)
		default:
			return nil, fmt.Errorf("Unhandled proto payload type: %T", proto)
		}
	default:
		event.SetSource(loggingSource)
		event.SetType(entry.Resource.Type)
	}
	logger.Debug("Created Stackdriver event: %v", event)
	return &event, nil
}
