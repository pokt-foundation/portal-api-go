package postgresdriver

import (
	"encoding/json"

	"github.com/lib/pq"
	"github.com/pokt-foundation/portal-api-go/repository"
)

type ListenerMock struct {
	Notify chan *pq.Notification
}

func NewListenerMock() *ListenerMock {
	return &ListenerMock{
		Notify: make(chan *pq.Notification, 32),
	}
}

func (l *ListenerMock) NotificationChannel() <-chan *pq.Notification {
	return l.Notify
}

func (l *ListenerMock) Listen(channel string) error {
	return nil
}

func gatewaySettingsIsNull(settings repository.GatewaySettings) bool {
	return settings.SecretKey == "" &&
		len(settings.WhitelistOrigins) == 0 &&
		len(settings.WhitelistUserAgents) == 0 &&
		len(settings.WhitelistContracts) == 0 &&
		len(settings.WhitelistMethods) == 0 &&
		len(settings.WhitelistBlockchains) == 0
}

func applicationInputs(mainTableAction, sideTablesAction Action, content any) []inputStruct {
	app := content.(*repository.Application)

	var inputs []inputStruct

	inputs = append(inputs, inputStruct{
		action: mainTableAction,
		table:  TableApplications,
		input: dbAppJSON{
			ApplicationID: app.ID,
			UserID:        app.UserID,
			Name:          app.Name,
			ContactEmail:  app.ContactEmail,
			Description:   app.Description,
			Owner:         app.Owner,
			URL:           app.URL,
			PayPlanType:   string(app.PayPlanType),
			Status:        string(app.Status),
			CreatedAt:     app.CreatedAt,
			UpdatedAt:     app.UpdatedAt,
		},
	})

	if app.GatewayAAT != (repository.GatewayAAT{}) {
		inputs = append(inputs, inputStruct{
			action: sideTablesAction,
			table:  TableGatewayAAT,
			input: dbGatewayAATJSON{
				ApplicationID:   app.ID,
				Address:         app.GatewayAAT.Address,
				ClientPublicKey: app.GatewayAAT.ClientPublicKey,
				PrivateKey:      app.GatewayAAT.PrivateKey,
				PublicKey:       app.GatewayAAT.ApplicationPublicKey,
				Signature:       app.GatewayAAT.ApplicationSignature,
				Version:         app.GatewayAAT.Version,
			},
		})
	}

	if !gatewaySettingsIsNull(app.GatewaySettings) {
		contracts, methods := marshalWhitelistContractsAndMethods(app.GatewaySettings.WhitelistContracts,
			app.GatewaySettings.WhitelistMethods)

		inputs = append(inputs, inputStruct{
			action: sideTablesAction,
			table:  TableGatewaySettings,
			input: dbGatewaySettingsJSON{
				ApplicationID:        app.ID,
				SecretKey:            app.GatewaySettings.SecretKey,
				SecretKeyRequired:    app.GatewaySettings.SecretKeyRequired,
				WhitelistContracts:   contracts,
				WhitelistMethods:     methods,
				WhitelistOrigins:     app.GatewaySettings.WhitelistOrigins,
				WhitelistUserAgents:  app.GatewaySettings.WhitelistUserAgents,
				WhitelistBlockchains: app.GatewaySettings.WhitelistBlockchains,
			},
		})
	}

	if app.NotificationSettings != (repository.NotificationSettings{}) {
		inputs = append(inputs, inputStruct{
			action: sideTablesAction,
			table:  TableNotificationSettings,
			input: dbNotificationSettingsJSON{
				ApplicationID: app.ID,
				SignedUp:      app.NotificationSettings.SignedUp,
				Quarter:       app.NotificationSettings.Quarter,
				Half:          app.NotificationSettings.Half,
				ThreeQuarters: app.NotificationSettings.ThreeQuarters,
				Full:          app.NotificationSettings.Full,
			},
		})
	}

	return inputs
}

func blockchainInputs(mainTableAction, sideTablesAction Action, content any) []inputStruct {
	blockchain := content.(*repository.Blockchain)

	var inputs []inputStruct

	inputs = append(inputs, inputStruct{
		action: mainTableAction,
		table:  TableBlockchains,
		input: dbBlockchainJSON{
			BlockchainID:      blockchain.ID,
			Altruist:          blockchain.Altruist,
			Blockchain:        blockchain.Blockchain,
			ChainID:           blockchain.ChainID,
			ChainIDCheck:      blockchain.ChainIDCheck,
			ChainPath:         blockchain.Path,
			Description:       blockchain.Description,
			EnforceResult:     blockchain.EnforceResult,
			Network:           blockchain.Network,
			Ticker:            blockchain.Ticker,
			BlockchainAliases: blockchain.BlockchainAliases,
			LogLimitBlocks:    blockchain.LogLimitBlocks,
			RequestTimeout:    blockchain.RequestTimeout,
			Active:            blockchain.Active,
			CreatedAt:         blockchain.CreatedAt,
			UpdatedAt:         blockchain.UpdatedAt,
		},
	})

	if blockchain.SyncCheckOptions != (repository.SyncCheckOptions{}) {
		inputs = append(inputs, inputStruct{
			action: sideTablesAction,
			table:  TableSyncCheckOptions,
			input: dbSyncCheckOptionsJSON{
				BlockchainID: blockchain.SyncCheckOptions.BlockchainID,
				Body:         blockchain.SyncCheckOptions.Body,
				Path:         blockchain.SyncCheckOptions.Path,
				ResultKey:    blockchain.SyncCheckOptions.ResultKey,
				Allowance:    blockchain.SyncCheckOptions.Allowance,
			},
		})
	}

	return inputs
}

func loadBalancerInputs(mainTableAction, sideTablesAction Action, content any) []inputStruct {
	lb := content.(*repository.LoadBalancer)

	var inputs []inputStruct

	inputs = append(inputs, inputStruct{
		action: mainTableAction,
		table:  TableLoadBalancers,
		input: dbLoadBalancerJSON{
			LbID:              lb.ID,
			Name:              lb.Name,
			UserID:            lb.UserID,
			RequestTimeout:    lb.RequestTimeout,
			Gigastake:         lb.Gigastake,
			GigastakeRedirect: lb.GigastakeRedirect,
			CreatedAt:         lb.CreatedAt,
			UpdatedAt:         lb.UpdatedAt,
		},
	})

	if !lb.StickyOptions.IsEmpty() {
		inputs = append(inputs, inputStruct{
			action: sideTablesAction,
			table:  TableStickinessOptions,
			input: dbStickinessOptionsJSON{
				LbID:       lb.ID,
				Duration:   lb.StickyOptions.Duration,
				Origins:    lb.StickyOptions.StickyOrigins,
				StickyMax:  lb.StickyOptions.StickyMax,
				Stickiness: lb.StickyOptions.Stickiness,
			},
		})
	}

	return inputs
}

func redirectInput(action Action, content any) inputStruct {
	redirect := content.(*repository.Redirect)

	return inputStruct{
		action: action,
		table:  TableRedirects,
		input: dbRedirectJSON{
			BlockchainID:   redirect.BlockchainID,
			Alias:          redirect.Alias,
			LoadBalancerID: redirect.LoadBalancerID,
			Domain:         redirect.Domain,
			CreatedAt:      redirect.CreatedAt,
			UpdatedAt:      redirect.UpdatedAt,
		},
	}
}

type inputStruct struct {
	action Action
	table  Table
	input  any
}

func mockInput(inStruct inputStruct) *pq.Notification {
	notification, _ := json.Marshal(Notification{
		Table:  inStruct.table,
		Action: inStruct.action,
		Data:   inStruct.input,
	})

	return &pq.Notification{
		Extra: string(notification),
	}
}

func mockContent(mainTableAction, sideTablesAction Action, content any) []*pq.Notification {
	var inputs []inputStruct

	switch content.(type) {
	case *repository.Application:
		inputs = applicationInputs(mainTableAction, sideTablesAction, content)
	case *repository.Blockchain:
		inputs = blockchainInputs(mainTableAction, sideTablesAction, content)
	case *repository.LoadBalancer:
		inputs = loadBalancerInputs(mainTableAction, sideTablesAction, content)
	case *repository.Redirect:
		inputs = []inputStruct{redirectInput(mainTableAction, content)}
	default:
		panic("type not supported")
	}

	var notifications []*pq.Notification

	for _, input := range inputs {
		notifications = append(notifications, mockInput(input))
	}

	return notifications
}

func (l *ListenerMock) MockEvent(mainTableAction, sideTablesAction Action, content any) {
	notifications := mockContent(mainTableAction, sideTablesAction, content)

	for _, notification := range notifications {
		l.Notify <- notification
	}
}
