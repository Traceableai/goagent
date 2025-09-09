package identifier

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
)

func TestSetServiceInstanceID(t *testing.T) {
	tests := []struct {
		name          string
		resourceAttrs map[string]string
		expectedID    string
		initialID     string
	}{
		{
			name:          "Nil resource attributes",
			resourceAttrs: nil,
			initialID:     "default-id",
			expectedID:    "default-id",
		},
		{
			name: "With service instance ID",
			resourceAttrs: map[string]string{
				ServiceInstanceIDKey: "test-instance-id",
			},
			initialID:  "default-id",
			expectedID: "test-instance-id",
		},
		{
			name: "Without service instance ID",
			resourceAttrs: map[string]string{
				"some.other.key": "some-value",
			},
			initialID:  "default-id",
			expectedID: "default-id",
		},
		{
			name: "Empty service instance ID",
			resourceAttrs: map[string]string{
				ServiceInstanceIDKey: "",
			},
			initialID:  "default-id",
			expectedID: "default-id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset to a known state for each test
			ServiceInstanceIDAttr = attribute.StringValue(tt.initialID)
			ServiceInstanceKeyValue = attribute.KeyValue{Key: ServiceInstanceIDKey, Value: ServiceInstanceIDAttr}

			// Call the function with test data
			SetServiceInstanceID(tt.resourceAttrs)

			// Verify the result
			assert.Equal(t, tt.expectedID, ServiceInstanceKeyValue.Value.AsString())
		})
	}
}
