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

	listenerMock.MockEvent(repository.ActionInsert, repository.ActionUpdate, &repository.Application{
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

	nMap := make(map[repository.Table]*repository.Notification, 4)

	n := <-driver.NotificationChannel()
	nMap[n.Table] = n

	n = <-driver.NotificationChannel()
	nMap[n.Table] = n

	n = <-driver.NotificationChannel()
	nMap[n.Table] = n

	n = <-driver.NotificationChannel()
	nMap[n.Table] = n

	c.Equal(repository.ActionInsert, nMap[repository.TableApplications].Action)
	app := nMap[repository.TableApplications].Data.(*repository.Application)
	c.Equal("321", app.ID)

	c.Equal(repository.ActionUpdate, nMap[repository.TableGatewayAAT].Action)
	aat := nMap[repository.TableGatewayAAT].Data.(*repository.GatewayAAT)
	c.Equal("321", aat.ID)
	c.Equal("123", aat.Address)

	c.Equal(repository.ActionUpdate, nMap[repository.TableGatewaySettings].Action)
	gSettings := nMap[repository.TableGatewaySettings].Data.(*repository.GatewaySettings)
	c.Equal("321", gSettings.ID)
	c.Equal("123", gSettings.SecretKey)

	c.Equal(repository.ActionUpdate, nMap[repository.TableNotificationSettings].Action)
	nSettings := nMap[repository.TableNotificationSettings].Data.(*repository.NotificationSettings)
	c.Equal("321", nSettings.ID)
	c.True(nSettings.Full)
}

func TestPostgresDriver_ListenBlockchain(t *testing.T) {
	c := require.New(t)

	listenerMock := NewListenerMock()
	driver := NewPostgresDriverFromSQLDBInstance(nil, listenerMock)

	listenerMock.MockEvent(repository.ActionInsert, repository.ActionUpdate, &repository.Blockchain{
		ID: "0021",
		SyncCheckOptions: repository.SyncCheckOptions{
			BlockchainID: "0021",
			Body:         "yeh",
		},
	})

	nMap := make(map[repository.Table]*repository.Notification, 2)

	n := <-driver.NotificationChannel()
	nMap[n.Table] = n

	n = <-driver.NotificationChannel()
	nMap[n.Table] = n

	c.Equal(repository.ActionInsert, nMap[repository.TableBlockchains].Action)
	blockchain := nMap[repository.TableBlockchains].Data.(*repository.Blockchain)
	c.Equal("0021", blockchain.ID)

	c.Equal(repository.ActionUpdate, nMap[repository.TableSyncCheckOptions].Action)
	syncOpts := nMap[repository.TableSyncCheckOptions].Data.(*repository.SyncCheckOptions)
	c.Equal("0021", syncOpts.BlockchainID)
	c.Equal("yeh", syncOpts.Body)
}

func TestPostgresDriver_ListenLoadBalancer(t *testing.T) {
	c := require.New(t)

	listenerMock := NewListenerMock()
	driver := NewPostgresDriverFromSQLDBInstance(nil, listenerMock)

	listenerMock.MockEvent(repository.ActionInsert, repository.ActionUpdate, &repository.LoadBalancer{
		ID: "123",
		StickyOptions: repository.StickyOptions{
			StickyOrigins: []string{"oahu"},
			Stickiness:    true,
		},
	})

	nMap := make(map[repository.Table]*repository.Notification, 2)

	n := <-driver.NotificationChannel()
	nMap[n.Table] = n

	n = <-driver.NotificationChannel()
	nMap[n.Table] = n

	c.Equal(repository.ActionInsert, nMap[repository.TableLoadBalancers].Action)
	lb := nMap[repository.TableLoadBalancers].Data.(*repository.LoadBalancer)
	c.Equal("123", lb.ID)

	c.Equal(repository.ActionUpdate, nMap[repository.TableStickinessOptions].Action)
	sOpts := nMap[repository.TableStickinessOptions].Data.(*repository.StickyOptions)
	c.Equal("123", sOpts.ID)
	c.Equal([]string{"oahu"}, sOpts.StickyOrigins)
	c.True(sOpts.Stickiness)
}

func TestPostgresDriver_ListenRedirect(t *testing.T) {
	c := require.New(t)

	listenerMock := NewListenerMock()
	driver := NewPostgresDriverFromSQLDBInstance(nil, listenerMock)

	listenerMock.MockEvent(repository.ActionInsert, repository.ActionUpdate, &repository.Redirect{
		BlockchainID: "0021",
	})

	n := <-driver.NotificationChannel()
	c.Equal(repository.ActionInsert, n.Action)
	c.Equal(repository.TableRedirects, n.Table)
	redirect := n.Data.(*repository.Redirect)
	c.Equal("0021", redirect.BlockchainID)
}

func TestPostgresDriver_ListenPanic(t *testing.T) {
	c := require.New(t)

	listenerMock := NewListenerMock()

	c.Panics(func() { listenerMock.MockEvent(repository.ActionInsert, repository.ActionUpdate, "dummy") })
}
