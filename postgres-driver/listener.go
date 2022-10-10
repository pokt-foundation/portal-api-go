package postgresdriver

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"
)

type Table string

const (
	TableApplications         Table = "applications"
	TableBlockchains          Table = "blockchains"
	TableGatewayAAT           Table = "gateway_aat"
	TableGatewaySettings      Table = "gateway_settings"
	TableLoadBalancers        Table = "loadbalancers"
	TableNotificationSettings Table = "notification_settings"
	TableRedirects            Table = "redirects"
	TableStickinessOptions    Table = "stickiness_options"
	TableSyncCheckOptions     Table = "sync_check_options"
)

type Action string

const (
	ActionInsert Action = "INSERT"
	ActionUpdate Action = "UPDATE"
)

type notification struct {
	Table  Table          `json:"table"`
	Action Action         `json:"action"`
	Data   map[string]any `json:"data"`
}

type Notification struct {
	Table  Table
	Action Action
	Data   any
}

type input interface {
	toOutput() any
}

func (n *notification) toNotificationWithType(inputType input) *Notification {
	rawData, _ := json.Marshal(n.Data)

	_ = json.Unmarshal(rawData, &inputType)

	return &Notification{
		Table:  n.Table,
		Action: n.Action,
		Data:   inputType.toOutput(),
	}
}

func (n *notification) toNotification() *Notification {
	switch n.Table {
	case TableApplications:
		return n.toNotificationWithType(dbAppJSON{})
	case TableBlockchains:
		return n.toNotificationWithType(dbBlockchainJSON{})
	case TableGatewayAAT:
		return n.toNotificationWithType(dbGatewayAATJSON{})
	case TableGatewaySettings:
		return n.toNotificationWithType(dbGatewaySettingsJSON{})
	case TableLoadBalancers:
		return n.toNotificationWithType(dbLoadBalancerJSON{})
	case TableNotificationSettings:
		return n.toNotificationWithType(dbNotificationSettingsJSON{})
	case TableRedirects:
		return n.toNotificationWithType(dbRedirectJSON{})
	case TableStickinessOptions:
		return n.toNotificationWithType(dbStickinessOptionsJSON{})
	case TableSyncCheckOptions:
		return n.toNotificationWithType(dbSyncCheckOptionsJSON{})
	}

	return nil
}

func (d *PostgresDriver) StartListener() Notification {
	reportProblem := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	listener := pq.NewListener(d.connString, 10*time.Second, time.Minute, reportProblem)
	err := listener.Listen("events")
	if err != nil {
		panic(err)
	}

	for {
		n := <-listener.Notify
		var notification notification
		err := json.Unmarshal([]byte(n.Extra), &notification)
		if err != nil {
			panic(err)
		}
		d.Notification <- notification.toNotification()
	}
}
