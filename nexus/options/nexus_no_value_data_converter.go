// The MIT License
//
// Copyright (c) 2020 Temporal Technologies Inc.  All rights reserved.
//
// Copyright (c) 2020 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package options

import (
	"fmt"
	"reflect"

	"github.com/nexus-rpc/sdk-go/nexus"
	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"
)

var (
	myDataConverter = converter.NewCompositeDataConverter(
		NewNexusNoValuePayloadConverter(),
		converter.NewNilPayloadConverter(),
		converter.NewByteSlicePayloadConverter(),

		// Order is important here. Both ProtoJsonPayload and ProtoPayload converters check for the same proto.Message
		// interface. The first match (ProtoJsonPayload in this case) will always be used for serialization.
		// Deserialization is controlled by metadata, therefore both converters can deserialize corresponding data format
		// (JSON or binary proto).
		converter.NewProtoJSONPayloadConverter(),
		converter.NewProtoPayloadConverter(),

		converter.NewJSONPayloadConverter(),
	)
)

// GetDefaultDataConverter returns default data converter used by Temporal worker.
func GetMyDataConverter() converter.DataConverter {
	return myDataConverter
}

const (
	// MetadataEncodingNil is "binary/nexusNoValue"
	MetadataEncodingNexusNoValue = "binary/nexusNoValue"
)

// NexusNoValuePayloadConverter doesn't set Data field in payload.
type NexusNoValuePayloadConverter struct {
}

// NewNexusNoValuePayloadConverter creates new instance of NexusNoValuePayloadConverter.
func NewNexusNoValuePayloadConverter() *NexusNoValuePayloadConverter {
	return &NexusNoValuePayloadConverter{}
}

// ToPayload converts single nil value to payload.
func (c *NexusNoValuePayloadConverter) ToPayload(value interface{}) (*commonpb.Payload, error) {
	if nexusNoValue, ok := value.(nexus.NoValue); ok {
		if nexusNoValue == nil {
			return newPayload(nil, c), nil
		}
	}
	return nil, nil
}

// FromPayload converts single nil value from payload.
func (c *NexusNoValuePayloadConverter) FromPayload(_ *commonpb.Payload, valuePtr interface{}) error {
	originalValue := reflect.ValueOf(valuePtr)
	if originalValue.Kind() != reflect.Ptr {
		return fmt.Errorf("type: %T: %w", valuePtr, converter.ErrValuePtrIsNotPointer)
	}

	originalValue = originalValue.Elem()
	if !originalValue.CanSet() {
		return fmt.Errorf("type: %T: %w", valuePtr, converter.ErrUnableToSetValue)
	}

	originalValue.Set(reflect.Zero(originalValue.Type()))
	return nil
}

// ToString converts payload object into human readable string.
func (c *NexusNoValuePayloadConverter) ToString(*commonpb.Payload) string {
	return "nexusNoValue"
}

// Encoding returns MetadataEncodingNil.
func (c *NexusNoValuePayloadConverter) Encoding() string {
	return MetadataEncodingNexusNoValue
}

func newPayload(data []byte, c converter.PayloadConverter) *commonpb.Payload {
	return &commonpb.Payload{
		Metadata: map[string][]byte{
			converter.MetadataEncoding: []byte(c.Encoding()),
		},
		Data: data,
	}
}
