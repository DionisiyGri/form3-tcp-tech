package request

import (
	"testing"
	"time"
)

const tolerance = 50 * time.Millisecond

func TestHandle(t *testing.T) {
	tests := []struct {
		name     string
		request  string
		expected string
		maxDelay time.Duration
	}{
		// Invalid requests
		{
			name:     "Invalid Request - Missing Delimiter",
			request:  "INVALIDREQUEST",
			expected: "RESPONSE|REJECTED|Invalid request",
			maxDelay: 0,
		},
		{
			name:     "Invalid Request - Incorrect Command",
			request:  "INVALID|123",
			expected: "RESPONSE|REJECTED|Invalid request",
			maxDelay: 0,
		},
		{
			name:     "Invalid Amount - Non-Integer",
			request:  "PAYMENT|abc",
			expected: "RESPONSE|REJECTED|Invalid amount",
			maxDelay: 0,
		},

		// Valid requests
		{
			name:     "Valid Payment - Amount <= 100",
			request:  "PAYMENT|50",
			expected: "RESPONSE|ACCEPTED|Transaction processed",
			maxDelay: 0,
		},
		{
			name:     "Valid Payment - Amount > 100",
			request:  "PAYMENT|150",
			expected: "RESPONSE|ACCEPTED|Transaction processed",
			maxDelay: 150 * time.Millisecond,
		},
		{
			name:     "Valid Payment - Amount > 10000",
			request:  "PAYMENT|15000",
			expected: "RESPONSE|ACCEPTED|Transaction processed",
			maxDelay: 10 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			result := Handle(tt.request)
			elapsed := time.Since(start)

			if result != tt.expected {
				t.Errorf("expected: %s, got: %s", tt.expected, result)
			}

			if tt.maxDelay > 0 {
				expectedMin := tt.maxDelay - tolerance
				expectedMax := tt.maxDelay + tolerance

				if elapsed < expectedMin || elapsed > expectedMax {
					t.Errorf("processing time out of bounds: expected %v ± %v, got %v", tt.maxDelay, tolerance, elapsed)
				}
			}
		})
	}
}
