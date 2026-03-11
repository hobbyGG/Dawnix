package client

import (
	"context"
	"testing"
)

func Test_emailClient_Send(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		to      string
		subject string
		body    string
		wantErr bool
	}{{
		name:    "valid email",
		to:      "hobbyGG@outlook.com",
		subject: "Test Email",
		body:    "This is a test email sent from the email client.",
		wantErr: false,
	},
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: construct the receiver type.
			gotErr := emailSender.Send(context.Background(), tt.to, tt.subject, tt.body)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Send() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Send() succeeded unexpectedly")
			}
		})
	}
}
