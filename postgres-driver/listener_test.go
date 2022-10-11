package postgresdriver

import (
	"testing"

	"github.com/pokt-foundation/portal-api-go/repository"
	"github.com/stretchr/testify/require"
)

func TestPostgresDriver_ListenApplication(t *testing.T) {
	c := require.New(t)

	listenerMock := NewListenerMock()
	driver := NewPostgresDriverFromSQLDBInstance(nil, listenerMock)

	listenerMock.MockEvent(ActionInsert, ActionUpdate, &repository.Application{
		ID: "321",
		GatewayAAT: repository.GatewayAAT{
			Address: "123",
		},
		GatewaySettings: repository.GatewaySettings{
			SecretKey: "123",
		},
		NotificationSettings: repository.NotificationSettings{
			Full: true,
		},
	})

	n := <-driver.Notification
	c.Equal(ActionInsert, n.Action)
	c.Equal(TableApplications, n.Table)
	app := n.Data.(*repository.Application)
	c.Equal("321", app.ID)

	n = <-driver.Notification
	c.Equal(ActionUpdate, n.Action)
	c.Equal(TableGatewayAAT, n.Table)
	aat := n.Data.(*repository.GatewayAAT)
	c.Equal("321", aat.ID)
	c.Equal("123", aat.Address)

	n = <-driver.Notification
	c.Equal(ActionUpdate, n.Action)
	c.Equal(TableGatewaySettings, n.Table)
	gSettings := n.Data.(*repository.GatewaySettings)
	c.Equal("321", gSettings.ID)
	c.Equal("123", gSettings.SecretKey)

	n = <-driver.Notification
	c.Equal(ActionUpdate, n.Action)
	c.Equal(TableNotificationSettings, n.Table)
	nSettings := n.Data.(*repository.NotificationSettings)
	c.Equal("321", nSettings.ID)
	c.True(nSettings.Full)
}

func TestPostgresDriver_ListenBlockchain(t *testing.T) {
	c := require.New(t)

	listenerMock := NewListenerMock()
	driver := NewPostgresDriverFromSQLDBInstance(nil, listenerMock)

	listenerMock.MockEvent(ActionInsert, ActionUpdate, &repository.Blockchain{
		ID: "0021",
		SyncCheckOptions: repository.SyncCheckOptions{
			BlockchainID: "0021",
			Body:         "yeh",
		},
	})

	n := <-driver.Notification
	c.Equal(ActionInsert, n.Action)
	c.Equal(TableBlockchains, n.Table)
	blockchain := n.Data.(*repository.Blockchain)
	c.Equal("0021", blockchain.ID)

	n = <-driver.Notification
	c.Equal(ActionUpdate, n.Action)
	c.Equal(TableSyncCheckOptions, n.Table)
	syncOpts := n.Data.(*repository.SyncCheckOptions)
	c.Equal("0021", syncOpts.BlockchainID)
	c.Equal("yeh", syncOpts.Body)
}

func TestPostgresDriver_ListenLoadBalancer(t *testing.T) {
	c := require.New(t)

	listenerMock := NewListenerMock()
	driver := NewPostgresDriverFromSQLDBInstance(nil, listenerMock)

	listenerMock.MockEvent(ActionInsert, ActionUpdate, &repository.LoadBalancer{
		ID: "123",
		StickyOptions: repository.StickyOptions{
			StickyOrigins: []string{"oahu"},
			Stickiness:    true,
		},
	})

	n := <-driver.Notification
	c.Equal(ActionInsert, n.Action)
	c.Equal(TableLoadBalancers, n.Table)
	lb := n.Data.(*repository.LoadBalancer)
	c.Equal("123", lb.ID)

	n = <-driver.Notification
	c.Equal(ActionUpdate, n.Action)
	c.Equal(TableStickinessOptions, n.Table)
	sOpts := n.Data.(*repository.StickyOptions)
	c.Equal("123", sOpts.ID)
	c.Equal([]string{"oahu"}, sOpts.StickyOrigins)
	c.True(sOpts.Stickiness)
}

func TestPostgresDriver_ListenRedirect(t *testing.T) {
	c := require.New(t)

	listenerMock := NewListenerMock()
	driver := NewPostgresDriverFromSQLDBInstance(nil, listenerMock)

	listenerMock.MockEvent(ActionInsert, ActionUpdate, &repository.Redirect{
		BlockchainID: "0021",
	})

	n := <-driver.Notification
	c.Equal(ActionInsert, n.Action)
	c.Equal(TableRedirects, n.Table)
	redirect := n.Data.(*repository.Redirect)
	c.Equal("0021", redirect.BlockchainID)
}

func TestPostgresDriver_ListenPanic(t *testing.T) {
	c := require.New(t)

	listenerMock := NewListenerMock()

	c.Panics(func() { listenerMock.MockEvent(ActionInsert, ActionUpdate, "dummy") })
}
