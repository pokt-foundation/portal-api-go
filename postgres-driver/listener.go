package postgresdriver

import (
	"encoding/json"

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

type Notification struct {
	Table  Table  `json:"table"`
	Action Action `json:"action"`
	Data   any    `json:"data"`
}

type Listener interface {
	NotificationChannel() <-chan *pq.Notification
	Listen(channel string) error
}

func (n *Notification) parseApplicationNotification() {
	rawData, _ := json.Marshal(n.Data)
	var dbApp dbAppJSON
	_ = json.Unmarshal(rawData, &dbApp)
	n.Data = dbApp.toOutput()
}

func (n *Notification) parseBlockchainNotification() {
	rawData, _ := json.Marshal(n.Data)
	var dbBlockchain dbBlockchainJSON
	_ = json.Unmarshal(rawData, &dbBlockchain)
	n.Data = dbBlockchain.toOutput()
}

func (n *Notification) parseGatewayAATNotification() {
	rawData, _ := json.Marshal(n.Data)
	var dbGatewayAAT dbGatewayAATJSON
	_ = json.Unmarshal(rawData, &dbGatewayAAT)
	n.Data = dbGatewayAAT.toOutput()
}

func (n *Notification) parseGatewaySettingsNotification() {
	rawData, _ := json.Marshal(n.Data)
	var dbGatewaySettings dbGatewaySettingsJSON
	_ = json.Unmarshal(rawData, &dbGatewaySettings)
	n.Data = dbGatewaySettings.toOutput()
}

func (n *Notification) parseLoadBalancerNotification() {
	rawData, _ := json.Marshal(n.Data)
	var dbLoadBalancer dbLoadBalancerJSON
	_ = json.Unmarshal(rawData, &dbLoadBalancer)
	n.Data = dbLoadBalancer.toOutput()
}

func (n *Notification) parseNotificationSettingsNotification() {
	rawData, _ := json.Marshal(n.Data)
	var dbNotificationSettings dbNotificationSettingsJSON
	_ = json.Unmarshal(rawData, &dbNotificationSettings)
	n.Data = dbNotificationSettings.toOutput()
}

func (n *Notification) parseRedirectNotification() {
	rawData, _ := json.Marshal(n.Data)
	var dbRedirect dbRedirectJSON
	_ = json.Unmarshal(rawData, &dbRedirect)
	n.Data = dbRedirect.toOutput()
}

func (n *Notification) parseStickinessOptionsNotification() {
	rawData, _ := json.Marshal(n.Data)
	var dbStickinessOpts dbStickinessOptionsJSON
	_ = json.Unmarshal(rawData, &dbStickinessOpts)
	n.Data = dbStickinessOpts.toOutput()
}

func (n *Notification) parseSyncOptionsNotification() {
	rawData, _ := json.Marshal(n.Data)
	var dbSyncOpts dbSyncCheckOptionsJSON
	_ = json.Unmarshal(rawData, &dbSyncOpts)
	n.Data = dbSyncOpts.toOutput()
}

func (n *Notification) parseNotification() {
	switch n.Table {
	case TableApplications:
		n.parseApplicationNotification()
	case TableBlockchains:
		n.parseBlockchainNotification()
	case TableGatewayAAT:
		n.parseGatewayAATNotification()
	case TableGatewaySettings:
		n.parseGatewaySettingsNotification()
	case TableLoadBalancers:
		n.parseLoadBalancerNotification()
	case TableNotificationSettings:
		n.parseNotificationSettingsNotification()
	case TableRedirects:
		n.parseRedirectNotification()
	case TableStickinessOptions:
		n.parseStickinessOptionsNotification()
	case TableSyncCheckOptions:
		n.parseSyncOptionsNotification()
	}
}

func (d *PostgresDriver) Listen() {
	for {
		n := <-d.listener.NotificationChannel()
		var notification Notification
		err := json.Unmarshal([]byte(n.Extra), &notification)
		if err != nil {
			panic(err)
		}
		notification.parseNotification()
		d.Notification <- &notification
	}
}
