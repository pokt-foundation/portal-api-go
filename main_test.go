package main

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	logger "github.com/sirupsen/logrus"
)

func TestGatherSettings(t *testing.T) {
	testCases := []struct {
		name        string
		args        []string
		expected    settings
		expectedErr error
	}{
		{
			name: "RPC URLs string",
			args: []string{"-rpcUrls", "https://url1,https://url2"},
			expected: settings{
				RPCURLs:  []string{"https://url1", "https://url2"},
				LogLevel: logger.InfoLevel,
				Port:     8090,
			},
		},
		{
			name: "Set all values",
			args: []string{
				"-rpcUrls", "https://url1",
				"-port", "8191",
				"-logLevel", "Debug",
				"-privateKey", "privateKey",
			},
			expected: settings{
				RPCURLs:    []string{"https://url1"},
				LogLevel:   logger.DebugLevel,
				Port:       8191,
				PrivateKey: "privateKey",
			},
		},
		{
			name:        "Invalid port number returns error",
			args:        []string{"-port", "foo"},
			expectedErr: fmt.Errorf("invalid value"),
		},
		{
			name:        "invalid arg returns error",
			args:        []string{"-invalid", "arg"},
			expectedErr: fmt.Errorf("flag provided but not defined"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := gatherSettings(tc.args)
			if tc.expectedErr != nil {
				if err == nil {
					t.Fatalf("Expected error: %v, got nil", tc.expectedErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.expected, actual); diff != "" {
				t.Errorf("unexpected value (-want +got):\n%s", diff)
			}
		})
	}
}
