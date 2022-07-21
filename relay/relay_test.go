package relay

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	logger "github.com/sirupsen/logrus"

	"github.com/pokt-foundation/pocket-go/provider"
	"github.com/pokt-foundation/pocket-go/relayer"

	"github.com/pokt-foundation/portal-api-go/repository"
	"github.com/pokt-foundation/portal-api-go/session"
	"github.com/pokt-foundation/portal-api-go/sticky"
)

var (
	app1 repository.Application = repository.Application{
		GatewayAAT: repository.GatewayAAT{
			Version: "gwaat_version",
		},
		FreeTierAAT: repository.FreeTierAAT{
			ClientPublicKey:      "free_aat_client_public_key",
			ApplicationPublicKey: "free_aat_app_public_key",
			ApplicationSignature: "free_aat_app_signature",
		},
	}
)

func TestSendRelay(t *testing.T) {
	testCases := []struct {
		name        string
		details     RelayDetails
		aatPlan     AatPlan
		relayError  error
		expected    []*relayer.Input
		expectedErr error
	}{
		{
			name: "relays with correct AAT",
			details: RelayDetails{
				Application: &app1,
			},
			aatPlan: AatPlanFreemium,
			expected: []*relayer.Input{
				{
					Method: "POST",
					PocketAAT: &provider.PocketAAT{
						Version:      "gwaat_version",
						AppPubKey:    "free_aat_app_public_key",
						ClientPubKey: "free_aat_client_public_key",
						Signature:    "free_aat_app_signature",
					},
					Node: &provider.Node{Address: "node-1"},
					Session: &provider.Session{
						Nodes: []*provider.Node{
							{Address: "node-1"},
						},
					},
				},
			},
		},
		{
			name: "nodeSticker is notified on failed relay",
			details: RelayDetails{
				Application: &app1,
			},
			relayError:  fmt.Errorf("relay failed"),
			expectedErr: fmt.Errorf("relay failed"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pocketRelayer := &fakePocketRelayer{relayError: tc.relayError}
			nodeSticker := &fakeNodeSticker{}
			rs := relayServer{
				log:            logger.New(),
				settings:       FreemiumSettings(),
				sessionManager: fakeSessionManager{},
				relayer:        pocketRelayer,
				nodeSticker:    nodeSticker,
			}

			err := rs.sendRelay(&tc.details)
			if tc.expectedErr != nil {
				if len(nodeSticker.failure) != 1 {
					t.Errorf("Expected node sticker service to have been notified %d time, found %d", 1, len(nodeSticker.failure))
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.expected, pocketRelayer.relays); diff != "" {
				t.Errorf("unexpected value (-want +got):\n%s", diff)
			}

			if tc.expectedErr == nil {
				if len(nodeSticker.success) != 1 {
					t.Errorf("Expected node sticker service to have been notified %d time, found %d", 1, len(nodeSticker.success))
				}
			}
		})
	}
}

func TestRelayWithLb(t *testing.T) {
	testCases := []struct {
		name        string
		lbs         []repository.LoadBalancer
		expectedErr error
	}{
		{
			name: "successfull relay",
		},
		{
			name: "relay to LB with no applications returns error",
			lbs:  []repository.LoadBalancer{},
		},
		// TODO: test case to verify an application is selected from the LB
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rs := relayServer{
				log:            logger.New(),
				settings:       FreemiumSettings(),
				sessionManager: fakeSessionManager{},
				relayer:        &fakePocketRelayer{},
				nodeSticker:    &fakeNodeSticker{},
			}

			d := RelayDetails{
				Application: &repository.Application{
					GatewayAAT: repository.GatewayAAT{
						Version: "gateway_version",
					},
					FreeTierAAT: repository.FreeTierAAT{
						ClientPublicKey:      "client_public_key",
						ApplicationPublicKey: "app_public_key",
						ApplicationSignature: "app_signature",
					},
				},
			}
			err := rs.sendRelay(&d)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}

}

type fakeSessionManager struct{}

func (f fakeSessionManager) GetSession(k session.Key) (*provider.Session, error) {
	return &provider.Session{
		Nodes: []*provider.Node{
			{
				Address: "node-1",
			},
		},
	}, nil
}

type fakePocketRelayer struct {
	relays     []*relayer.Input
	relayError error
}

func (f *fakePocketRelayer) Relay(input *relayer.Input, options *provider.RelayRequestOptions) (*relayer.Output, error) {
	f.relays = append(f.relays, input)
	return &relayer.Output{}, f.relayError
}

type fakeRepository struct {
	lbs []repository.LoadBalancer
}

func (f fakeRepository) GetLoadBalancer(id string) (repository.LoadBalancer, error) {
	for _, lb := range f.lbs {
		if lb.ID == id {
			return lb, nil
		}
	}
	return repository.LoadBalancer{}, fmt.Errorf("LoadBalancer not found")
}

type fakeNodeSticker struct {
	success []*sticky.StickyDetails
	failure []*sticky.StickyDetails
}

func (f *fakeNodeSticker) GetStickyDetails(repository.StickyOptions, sticky.KeyBuilder, sticky.OptionsVerifier) sticky.StickyDetails {
	return sticky.StickyDetails{}
}

func (f *fakeNodeSticker) Success(d *sticky.StickyDetails) error {
	f.success = append(f.success, d)
	return nil
}

func (f *fakeNodeSticker) Failure(d *sticky.StickyDetails) error {
	f.failure = append(f.failure, d)
	return nil
}
