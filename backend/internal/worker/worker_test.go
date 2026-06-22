package worker

import "testing"

func TestNextRetryAttempt(t *testing.T) {
	tests := []struct {
		name       string
		attempt    int
		maxRetries int
		next       int
		retry      bool
	}{
		{name: "first retry", attempt: 0, maxRetries: 3, next: 1, retry: true},
		{name: "last retry", attempt: 2, maxRetries: 3, next: 3, retry: true},
		{name: "retry limit reached", attempt: 3, maxRetries: 3, next: 3, retry: false},
		{name: "default retry limit", attempt: 2, maxRetries: 0, next: 3, retry: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			next, retry := nextRetryAttempt(test.attempt, test.maxRetries)
			if next != test.next || retry != test.retry {
				t.Fatalf("expected (%d, %t), got (%d, %t)", test.next, test.retry, next, retry)
			}
		})
	}
}
