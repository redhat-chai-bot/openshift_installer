package quota

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/api/googleapi"
)

func TestHandleGCPQuotaLoadError(t *testing.T) {
	const serviceUnavailableMessage = "The service is currently unavailable."
	serviceUnavailable := &googleapi.Error{
		Code:    http.StatusServiceUnavailable,
		Message: serviceUnavailableMessage,
		Errors: []googleapi.ErrorItem{{
			Reason:  "backendError",
			Message: serviceUnavailableMessage,
		}},
	}
	assert.Equal(t, "googleapi: Error 503: The service is currently unavailable., backendError", serviceUnavailable.Error())

	tests := []struct {
		name     string
		err      error
		wantSkip bool
		wantErr  string
	}{
		{
			name:     "service usage backend error is nonfatal",
			err:      fmt.Errorf("failed to load quota limits: %w", serviceUnavailable),
			wantSkip: true,
		},
		{
			name: "non-transient error is fatal",
			err: &googleapi.Error{
				Code:    http.StatusBadRequest,
				Message: "bad request",
			},
			wantErr: "failed to load Quota for services: compute.googleapis.com, iam.googleapis.com: googleapi: Error 400: bad request",
		},
		{
			name: "successful load continues quota check",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			skip, err := handleGCPQuotaLoadError(test.err, []string{"compute.googleapis.com", "iam.googleapis.com"})

			if test.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, test.wantErr)
			}
			assert.Equal(t, test.wantSkip, skip)
		})
	}
}
