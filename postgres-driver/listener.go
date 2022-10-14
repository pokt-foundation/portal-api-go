package postgresdriver

import (
	"encoding/json"

	"github.com/lib/pq"
	"github.com/pokt-foundation/portal-api-go/repository"
)

type Listener interface {
	NotificationChannel() <-chan *pq.Notification
	Listen(channel string) error
}

func parseApplicationNotification(n *repository.Notification) {
	rawData, _ := json.Marshal(n.Data)
	var dbApp dbAppJSON
	_ = json.Unmarshal(rawData, &dbApp)
	n.Data = dbApp.toOutput()
}

func parseBlockchainNotification(n *repository.Notification) {
	rawData, _ := json.Marshal(n.Data)
	var dbBlockchain dbBlockchainJSON
	_ = json.Unmarshal(rawData, &dbBlockchain)
	n.Data = dbBlockchain.toOutput()
}

func parseGatewayAATNotification(n *repository.Notification) {
	rawData, _ := json.Marshal(n.Data)
	var dbGatewayAAT dbGatewayAATJSON
	_ = json.Unmarshal(rawData, &dbGatewayAAT)
	n.Data = dbGatewayAAT.toOutput()
}

func parseGatewaySettingsNotification(n *repository.Notification) {
	rawData, _ := json.Marshal(n.Data)
	var dbGatewaySettings dbGatewaySettingsJSON
	_ = json.Unmarshal(rawData, &dbGatewaySettings)
	n.Data = dbGatewaySettings.toOutput()
}

func parseLoadBalancerNotification(n *repository.Notification) {
	rawData, _ := json.Marshal(n.Data)
	var dbLoadBalancer dbLoadBalancerJSON
	_ = json.Unmarshal(rawData, &dbLoadBalancer)
	n.Data = dbLoadBalancer.toOutput()
}

func parseNotificationSettingsNotification(n *repository.Notification) {
	rawData, _ := json.Marshal(n.Data)
	var dbNotificationSettings dbNotificationSettingsJSON
	_ = json.Unmarshal(rawData, &dbNotificationSettings)
	n.Data = dbNotificationSettings.toOutput()
}

func parseRedirectNotification(n *repository.Notification) {
	rawData, _ := json.Marshal(n.Data)
	var dbRedirect dbRedirectJSON
	_ = json.Unmarshal(rawData, &dbRedirect)
	n.Data = dbRedirect.toOutput()
}

func parseStickinessOptionsNotification(n *repository.Notification) {
	rawData, _ := json.Marshal(n.Data)
	var dbStickinessOpts dbStickinessOptionsJSON
	_ = json.Unmarshal(rawData, &dbStickinessOpts)
	n.Data = dbStickinessOpts.toOutput()
}

func parseSyncOptionsNotification(n *repository.Notification) {
	rawData, _ := json.Marshal(n.Data)
	var dbSyncOpts dbSyncCheckOptionsJSON
	_ = json.Unmarshal(rawData, &dbSyncOpts)
	n.Data = dbSyncOpts.toOutput()
}

func parseNotification(n *repository.Notification) {
	switch n.Table {
	case repository.TableApplications:
		parseApplicationNotification(n)
	case repository.TableBlockchains:
		parseBlockchainNotification(n)
	case repository.TableGatewayAAT:
		parseGatewayAATNotification(n)
	case repository.TableGatewaySettings:
		parseGatewaySettingsNotification(n)
	case repository.TableLoadBalancers:
		parseLoadBalancerNotification(n)
	case repository.TableNotificationSettings:
		parseNotificationSettingsNotification(n)
	case repository.TableRedirects:
		parseRedirectNotification(n)
	case repository.TableStickinessOptions:
		parseStickinessOptionsNotification(n)
	case repository.TableSyncCheckOptions:
		parseSyncOptionsNotification(n)
	}
}

func parsePQNotification(n *pq.Notification, outCh chan *repository.Notification) {
	if n != nil {
		var notification repository.Notification
		_ = json.Unmarshal([]byte(n.Extra), &notification)
		parseNotification(&notification)
		outCh <- &notification
	}
}

func Listen(inCh <-chan *pq.Notification, outCh chan *repository.Notification) {
	for {
		n := <-inCh
		go parsePQNotification(n, outCh)
	}
}
