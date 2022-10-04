package web

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"
	logger "github.com/sirupsen/logrus"

	"github.com/pokt-foundation/portal-api-go/relay"
)

func TestGetIDs(t *testing.T) {
	testCases := []struct {
		name         string
		path         string
		expectedApp  string
		expectedLb   string
		expectedPath string
		expectedErr  error
	}{
		{
			name:        "App ID is extracted",
			path:        "/v1/app-1234567890",
			expectedApp: "app-1234567890",
		},
		{
			name:       "LB ID is extracted",
			path:       "/v1/lB/lb-1234567890",
			expectedLb: "lb-1234567890",
		},
		{
			name:        "Invalid path is rejected",
			path:        "/invalid-path",
			expectedErr: ErrInvalidPath,
		},
		{
			name:        "App ID is truncated",
			path:        "/v1/app-123456789012345678901234567890",
			expectedApp: "app-12345678901234567890",
		},
		{
			name:       "LB ID is truncated",
			path:       "/v1/lb/lb-123456789012345678901234567890",
			expectedLb: "lb-123456789012345678901",
		},
		{
			name:         "Relay path is extracted from App ID",
			path:         "/v1/lb/lb-123456789012345678901~relayPath~001",
			expectedLb:   "lb-123456789012345678901",
			expectedPath: "/relayPath/001",
		},
		{
			name:         "Relay path is extracted from Lb ID",
			path:         "/v1/app-12345678901234567890~relayPath~001",
			expectedApp:  "app-12345678901234567890",
			expectedPath: "/relayPath/001",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			app, lb, path, err := ids("eth-mainnet.pokt.network" + tc.path)
			if tc.expectedErr != nil {
				if !errors.Is(err, tc.expectedErr) {
					t.Fatalf("Expected err: %v, got: %v", tc.expectedErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if tc.expectedApp != app {
				t.Errorf("Expected appID: %q, got: %q", tc.expectedApp, app)
			}
			if tc.expectedLb != lb {
				t.Errorf("Expected lbID: %q, got: %q", tc.expectedLb, lb)
			}
			if tc.expectedPath != path {
				t.Errorf("Expected relayPath: %q, got: %q", tc.expectedPath, path)
			}
		})
	}
}

func TestBuildRelayOptions(t *testing.T) {
	testCases := []struct {
		name        string
		req         *http.Request
		expected    relay.RelayOptions
		expectedErr error
	}{
		{
			name: "valid http request on loadbalancer endpoint",
			req: &http.Request{
				URL: &url.URL{
					Path: "eth-mainnet.pokt.network/v1/lb/lb-123456789012345678901~relay~path~12",
				},
				Body: ioutil.NopCloser(bytes.NewReader(
					[]byte(`{"blockchainID": "0001", "rawData": {"method": "post", "rpcID": "rpcID002"}}`),
				)),
				Header: map[string][]string{
					"Origin": {"origin-foo"},
				},
			},
			expected: relay.RelayOptions{
				Path:           "/relay/path/12",
				Method:         "POST",
				Host:           "eth-mainnet",
				LoadBalancerID: "lb-123456789012345678901",
				Origin:         "origin-foo",
				BlockchainID:   "0001",
				RawData:        string(`{"method":"post","rpcID":"rpcID002"}`),
			},
		},
		{
			name: "Invalid request path",
			req: &http.Request{
				URL: &url.URL{
					Path: "/invalid-path",
				},
				Body: ioutil.NopCloser(bytes.NewReader([]byte(`{"body": "body"}`))), // &httpRequestBody{[]byte("body")},
			},
			expectedErr: ErrInvalidPath,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := buildRelayOptions(tc.req)
			if tc.expectedErr != nil {
				if err == nil {
					t.Fatalf("Expected error: %v, got nil", tc.expectedErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			tc.expected.RequestID = actual.RequestID
			if diff := cmp.Diff(tc.expected, actual); diff != "" {
				t.Errorf("unexpected value (-want +got):\n%s", diff)
			}
		})
	}
}

func TestGetHttpServer(t *testing.T) {
	expectedRelay := relay.RelayOptions{
		Method:       http.MethodPost,
		Host:         "eth-mainnet",
		Origin:       "origin-foo",
		Path:         "/relay/path/12",
		RawData:      string(`{"method":"post","rpcID":"rpcID002"}`),
		BlockchainID: "0001",
	}

	testCases := []struct {
		name          string
		path          string
		expectedAppID string
		expectedLbID  string
		expectedErr   error
	}{
		{
			name:         "Relay with Load Balancer is sent to correct relayer handler",
			path:         "eth-mainnet.pokt.network/v1/lb/lb-123456789012345678901~relay~path~12",
			expectedLbID: "lb-123456789012345678901",
		},
		{
			name:          "Relay with Application is sent to correct relayer handler",
			path:          "eth-mainnet.pokt.network/v1/app-12345678901234567890~relay~path~12",
			expectedAppID: "app-12345678901234567890",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			f := fakeRelayer{}
			httpServer := GetHttpServer(&f, logger.New())
			//TODO: use httptest.NewRequest
			req := &http.Request{
				Method: http.MethodPost,
				URL:    &url.URL{Path: tc.path},
				Body: ioutil.NopCloser(bytes.NewReader(
					[]byte(`{"blockchainID": "0001", "rawData": {"method": "post", "rpcID": "rpcID002"}}`),
				)),
				Header: map[string][]string{
					"Origin": {"origin-foo"},
				},
			}

			w := httptest.NewRecorder()
			httpServer(w, req)
			resp := w.Result()
			if resp.StatusCode != 200 {
				t.Fatalf("Expected status code: 200, got: %d", resp.StatusCode)
			}

			expected := expectedRelay
			var actual relay.RelayOptions
			if tc.expectedLbID != "" {
				expected.LoadBalancerID = tc.expectedLbID
				actual = f.lbRelay
			} else if tc.expectedAppID != "" {
				expected.ApplicationID = tc.expectedAppID
				actual = f.appRelay
			}
			// Copy UUID as it cannot be set deterministically
			expected.RequestID = actual.RequestID
			if diff := cmp.Diff(expected, actual); diff != "" {
				t.Errorf("unexpected LB relay value (-want +got):\n%s", diff)
			}
		})
	}
}

type fakeRelayer struct {
	appRelay relay.RelayOptions
	lbRelay  relay.RelayOptions
}

func (f *fakeRelayer) RelayWithApp(r relay.RelayOptions) error {
	f.appRelay = r
	return nil
}

func (f *fakeRelayer) RelayWithLb(r relay.RelayOptions) error {
	f.lbRelay = r
	return nil
}
