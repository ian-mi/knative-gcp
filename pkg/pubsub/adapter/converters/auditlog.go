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
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/empty"
	auditpb "google.golang.org/genproto/googleapis/cloud/audit"
	logpb "google.golang.org/genproto/googleapis/logging/v2"
)

const (
	AuditLogAdapterType = "google.auditlog"

	logEntrySchema = "type.googleapis.com/google.logging.v2.LogEntry"
	loggingSource  = "logging.googleapis.com"
	EventType      = "com.google.cloud.auditlog.event"
)

var (
	jsonpbUnmarshaller = jsonpb.Unmarshaler{
		AllowUnknownFields: true,
		AnyResolver:        resolver(resolveAnyUnknowns),
	}
	jsonpbMarshaler = jsonpb.Marshaler{}
)

// Resolver function type that can be used to resolve Any fields in a jsonpb.Unmarshaler.
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

// Resolver type which resolves unknown message types to empty.Empty.
func resolveAnyUnknowns(typeURL string) (proto.Message, error) {
	// Only the part of typeUrl after the last slash is relevant.
	mname := typeURL
	if slash := strings.LastIndex(mname, "/"); slash >= 0 {
		mname = mname[slash+1:]
	}
	mt := proto.MessageType(mname)
	if mt == nil {
		return (*UnknownMsg)(&empty.Empty{}), nil
	}
	return reflect.New(mt.Elem()).Interface().(proto.Message), nil
}

func convertAuditLog(ctx context.Context, msg *cepubsub.Message, sendMode ModeType) (*cloudevents.Event, error) {
	if msg == nil {
		return nil, fmt.Errorf("nil pubsub message")
	}
	entry := logpb.LogEntry{}
	if err := jsonpbUnmarshaller.Unmarshal(bytes.NewReader(msg.Data), &entry); err != nil {
		return nil, fmt.Errorf("failed to decode LogEntry: %q", err)
	}

	// Make a new event and convert the message payload.
	event := cloudevents.NewEvent(cloudevents.VersionV1)
	event.SetID(entry.InsertId + entry.LogName + ptypes.TimestampString(entry.Timestamp))
	if timestamp, err := ptypes.Timestamp(entry.Timestamp); err != nil {
		return nil, fmt.Errorf("invalid LogEntry timestamp: %q", err)
	} else {
		event.SetTime(timestamp)
	}
	event.SetData(msg.Data)
	event.SetDataSchema(logEntrySchema)
	event.SetDataContentType(cloudevents.ApplicationJSON)

	switch payload := entry.Payload.(type) {
	case *logpb.LogEntry_ProtoPayload:
		var unpacked ptypes.DynamicAny
		if err := ptypes.UnmarshalAny(payload.ProtoPayload, &unpacked); err != nil {
			return nil, fmt.Errorf("unrecognized proto payload: %q", err)
		}
		switch proto := unpacked.Message.(type) {
		case *auditpb.AuditLog:
			event.SetSource(proto.ServiceName)
			event.SetSubject(proto.MethodName)
			event.SetType(EventType)
		default:
			return nil, fmt.Errorf("unhandled proto payload type: %T", proto)
		}
	default:
		return nil, fmt.Errorf("non-AuditLog log entry")
	}
	return &event, nil
}