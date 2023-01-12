package postgresdriver

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/pokt-foundation/portal-api-go/repository"
)

func TestListen(t *testing.T) {
	testCases := []struct {
		name                  string
		content               repository.SavedOnDB
		expectedNotifications map[repository.Table]*repository.Notification
		wantPanic             bool
	}{
		{
			name: "application",
			content: &repository.Application{
				ID: "321",
				GatewayAAT: repository.GatewayAAT{
					Address: "123",
				},
				GatewaySettings: repository.GatewaySettings{
					SecretKey: "123",
					WhitelistContracts: []repository.WhitelistContract{
						{BlockchainID: "001", Contracts: []string{"test123", "test456"}},
					},
					WhitelistMethods: []repository.WhitelistMethod{
						{BlockchainID: "001", Methods: []string{"POST"}},
					},
				},
				Limit: repository.AppLimit{
					PayPlan:     repository.PayPlan{Type: repository.Enterprise},
					CustomLimit: 2000000,
				},
				NotificationSettings: repository.NotificationSettings{
					Full: true,
				},
			},
			expectedNotifications: map[repository.Table]*repository.Notification{
				repository.TableApplications: {
					Table:  repository.TableApplications,
					Action: repository.ActionInsert,
					Data: &repository.Application{
						ID: "321",
					},
				},
				repository.TableGatewayAAT: {
					Table:  repository.TableGatewayAAT,
					Action: repository.ActionUpdate,
					Data: &repository.GatewayAAT{
						ID:      "321",
						Address: "123",
					},
				},
				repository.TableAppLimits: {
					Table:  repository.TableAppLimits,
					Action: repository.ActionUpdate,
					Data: &repository.AppLimit{
						ID:          "321",
						PayPlan:     repository.PayPlan{Type: repository.Enterprise, Limit: 0},
						CustomLimit: 2000000,
					},
				},
				repository.TableGatewaySettings: {
					Table:  repository.TableGatewaySettings,
					Action: repository.ActionUpdate,
					Data: &repository.GatewaySettings{
						ID:        "321",
						SecretKey: "123",
					},
				},
				repository.TableWhitelistContracts: {
					Table:  repository.TableWhitelistContracts,
					Action: repository.ActionUpdate,
					Data: &repository.WhitelistContract{
						ID: "321", BlockchainID: "001", Contracts: []string{"test123", "test456"},
					},
				},
				repository.TableWhitelistMethods: {
					Table:  repository.TableWhitelistMethods,
					Action: repository.ActionUpdate,
					Data: &repository.WhitelistMethod{
						ID: "321", BlockchainID: "001", Methods: []string{"POST"},
					},
				},
				repository.TableNotificationSettings: {
					Table:  repository.TableNotificationSettings,
					Action: repository.ActionUpdate,
					Data: &repository.NotificationSettings{
						ID:   "321",
						Full: true,
					},
				},
			},
		},
		{
			name: "blockchain",
			content: &repository.Blockchain{
				ID: "0021",
				SyncCheckOptions: repository.SyncCheckOptions{
					BlockchainID: "0021",
					Body:         "yeh",
				},
			},
			expectedNotifications: map[repository.Table]*repository.Notification{
				repository.TableBlockchains: {
					Table:  repository.TableBlockchains,
					Action: repository.ActionInsert,
					Data: &repository.Blockchain{
						ID: "0021",
					},
				},
				repository.TableSyncCheckOptions: {
					Table:  repository.TableSyncCheckOptions,
					Action: repository.ActionUpdate,
					Data: &repository.SyncCheckOptions{
						BlockchainID: "0021",
						Body:         "yeh",
					},
				},
			},
		},
		{
			name: "load balancer",
			content: &repository.LoadBalancer{
				ID: "123",
				StickyOptions: repository.StickyOptions{
					StickyOrigins: []string{"oahu"},
					Stickiness:    true,
				},
				ApplicationIDs: []string{"a123"},
			},
			expectedNotifications: map[repository.Table]*repository.Notification{
				repository.TableLoadBalancers: {
					Table:  repository.TableLoadBalancers,
					Action: repository.ActionInsert,
					Data: &repository.LoadBalancer{
						ID: "123",
					},
				},
				repository.TableStickinessOptions: {
					Table:  repository.TableStickinessOptions,
					Action: repository.ActionUpdate,
					Data: &repository.StickyOptions{
						ID:            "123",
						StickyOrigins: []string{"oahu"},
						Stickiness:    true,
					},
				},
				repository.TableLbApps: {
					Table:  repository.TableLbApps,
					Action: repository.ActionUpdate,
					Data: &repository.LbApp{
						LbID:  "123",
						AppID: "a123",
					},
				},
			},
		},
		{
			name: "redirect",
			content: &repository.Redirect{
				BlockchainID: "0021",
			},
			expectedNotifications: map[repository.Table]*repository.Notification{
				repository.TableRedirects: {
					Table:  repository.TableRedirects,
					Action: repository.ActionInsert,
					Data: &repository.Redirect{
						BlockchainID: "0021",
					},
				},
			},
		},
		{
			name:      "panic",
			content:   &repository.GatewayAAT{},
			wantPanic: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if (r != nil) != tc.wantPanic {
					t.Errorf("recover = %v, wantPanic = %v", r, tc.wantPanic)
				}
			}()

			listenerMock := NewListenerMock()
			driver := NewPostgresDriverFromSQLDBInstance(nil, listenerMock)

			listenerMock.MockEvent(repository.ActionInsert, repository.ActionUpdate, tc.content)

			time.Sleep(1 * time.Second)
			driver.CloseListener()

			nMap := make(map[repository.Table]*repository.Notification)

			for n := range driver.NotificationChannel() {
				nMap[n.Table] = n
			}

			if diff := cmp.Diff(tc.expectedNotifications, nMap); diff != "" {
				t.Errorf("unexpected value (-want +got):\n%s", diff)
			}
		})
	}
}
