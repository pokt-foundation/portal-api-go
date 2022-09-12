package postgresdriver

import (
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/pokt-foundation/portal-api-go/repository"
	"github.com/stretchr/testify/require"
)

func TestPostgresDriver_ReadApplications(t *testing.T) {
	c := require.New(t)

	db, mock, err := sqlmock.New()
	c.NoError(err)

	defer db.Close()

	rows := sqlmock.NewRows([]string{"application_id", "contact_email", "created_at", "description",
		"name", "owner", "updated_at", "url", "user_id", "whitelist_contracts", "whitelist_methods"}).
		AddRow("5f62b7d8be3591c4dea85661", "dummy@ocampoent.com", time.Now(), "Wawawa gateway",
			"Wawawa", "ohana", time.Now(), "https://dummy.com", "6068da279aab4900333ec6dd",
			`[{"blockchainID":"0021","contracts":["0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48"]}]`,
			`[{"blockchainID":"000C","methods":["\t  eth_getBlockByHash"]}]`)

	mock.ExpectQuery("^SELECT (.+) FROM applications (.+)").WillReturnRows(rows)

	driver := NewPostgresDriverFromSQLDBInstance(db)

	applications, err := driver.ReadApplications()
	c.NoError(err)
	c.Len(applications, 1)

	mock.ExpectQuery("^SELECT (.+) FROM applications (.+)").WillReturnError(errors.New("dummy error"))

	applications, err = driver.ReadApplications()
	c.EqualError(err, "dummy error")
	c.Empty(applications)
}

func TestPostgresDriver_WriteApplication(t *testing.T) {
	c := require.New(t)

	db, mock, err := sqlmock.New()
	c.NoError(err)

	defer db.Close()

	driver := NewPostgresDriverFromSQLDBInstance(db)

	mock.ExpectBegin()

	mock.ExpectExec("INSERT into applications").WithArgs(sqlmock.AnyArg(),
		"60ddc61b6e29c3003378361D", "klk", "yes@yes.com", "a life", "juancito", "app.com", "FREETIER_V0", "ORPHANED", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("INSERT into gateway_aat").WithArgs(sqlmock.AnyArg(),
		"f463b4dd88d865c22acbf38981b6c505bcc46c64",
		"8aaedb01a840fd6c9ab5019786c485bd98e69ca492cdb685aabee8473e7fad77",
		sqlmock.AnyArg(),
		"7a80a331c20cb0ac9e30a0d0c68df5f334b9c8bbe10dcfd95b6cb42bf412037d",
		"1566702d9a667c6007639eeb47a48cd2fed79592c5db9040eadc89f81748a4adef82711854b32065ddafa86eec1f1ed3b6f4f03a0786d9cf12c5262b948d9c01",
		"1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("INSERT into public_pocket_account").WithArgs(sqlmock.AnyArg(),
		"7a80a331c20cb0ac9e30a0d0c68df5f334b9c8bbe10dcfd95b6cb42bf412037d", "f463b4dd88d865c22acbf38981b6c505bcc46c64").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("INSERT into gateway_settings").WithArgs(sqlmock.AnyArg(),
		"54y4p93body6qco2nrhonz6bltn1k5e8", true, `[{"blockchainID":"0021","contracts":["ajua"]}]`,
		`[{"BlockchainID":"0021","methods":["POST"]}]`, pq.StringArray([]string{"url.com"}), pq.StringArray([]string{"gecko.com"}), pq.StringArray([]string{"0021"})).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("INSERT into notification_settings").WithArgs(sqlmock.AnyArg(),
		true, true, true, true, true).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	appToSend := &repository.Application{
		ID:           "60ddc61b6e29c3003378361D",
		UserID:       "60ddc61b6e29c3003378361D",
		Name:         "klk",
		ContactEmail: "yes@yes.com",
		Description:  "a life",
		Owner:        "juancito",
		URL:          "app.com",
		PayPlanType:  repository.FreetierV0,
		Status:       repository.Orphaned,
		GatewayAAT: repository.GatewayAAT{
			ApplicationPublicKey: "7a80a331c20cb0ac9e30a0d0c68df5f334b9c8bbe10dcfd95b6cb42bf412037d",
			PrivateKey:           "",
			Address:              "f463b4dd88d865c22acbf38981b6c505bcc46c64",
			ApplicationSignature: "1566702d9a667c6007639eeb47a48cd2fed79592c5db9040eadc89f81748a4adef82711854b32065ddafa86eec1f1ed3b6f4f03a0786d9cf12c5262b948d9c01",
			ClientPublicKey:      "8aaedb01a840fd6c9ab5019786c485bd98e69ca492cdb685aabee8473e7fad77",
			Version:              "1",
		},
		PublicPocketAccount: repository.PublicPocketAccount{
			PublicKey: "7a80a331c20cb0ac9e30a0d0c68df5f334b9c8bbe10dcfd95b6cb42bf412037d",
			Address:   "f463b4dd88d865c22acbf38981b6c505bcc46c64",
		},
		GatewaySettings: repository.GatewaySettings{
			SecretKey:         "54y4p93body6qco2nrhonz6bltn1k5e8",
			SecretKeyRequired: true,
			WhitelistContracts: []repository.WhitelistContract{
				{
					BlockchainID: "0021",
					Contracts:    []string{"ajua"},
				},
			},
			WhitelistMethods: []repository.WhitelistMethod{
				{
					BlockchainID: "0021",
					Methods:      []string{"POST"},
				},
			},
			WhitelistOrigins:     []string{"url.com"},
			WhitelistUserAgents:  []string{"gecko.com"},
			WhitelistBlockchains: []string{"0021"},
		},
		NotificationSettings: repository.NotificationSettings{
			SignedUp:      true,
			Quarter:       true,
			Half:          true,
			ThreeQuarters: true,
			Full:          true,
		},
	}

	app, err := driver.WriteApplication(appToSend)
	c.NoError(err)
	c.NotEmpty(app.ID)

	mock.ExpectBegin()

	mock.ExpectExec("INSERT into applications").WithArgs(sqlmock.AnyArg(),
		"60ddc61b6e29c3003378361D", "klk", "yes@yes.com", "a life", "juancito", "app.com", "FREETIER_V0", "ORPHANED", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(errors.New("error in applications"))

	app, err = driver.WriteApplication(appToSend)
	c.EqualError(err, "error in applications")
	c.Empty(app)

	mock.ExpectBegin()

	mock.ExpectExec("INSERT into applications").WithArgs(sqlmock.AnyArg(),
		"60ddc61b6e29c3003378361D", "klk", "yes@yes.com", "a life", "juancito", "app.com", "FREETIER_V0", "ORPHANED", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("INSERT into gateway_aat").WithArgs(sqlmock.AnyArg(),
		"f463b4dd88d865c22acbf38981b6c505bcc46c64",
		"8aaedb01a840fd6c9ab5019786c485bd98e69ca492cdb685aabee8473e7fad77",
		sqlmock.AnyArg(),
		"7a80a331c20cb0ac9e30a0d0c68df5f334b9c8bbe10dcfd95b6cb42bf412037d",
		"1566702d9a667c6007639eeb47a48cd2fed79592c5db9040eadc89f81748a4adef82711854b32065ddafa86eec1f1ed3b6f4f03a0786d9cf12c5262b948d9c01",
		"1").
		WillReturnError(errors.New("error in gateway_aat"))

	app, err = driver.WriteApplication(appToSend)
	c.EqualError(err, "error in gateway_aat")
	c.Empty(app)

	appToSend.Status = "wrong"

	app, err = driver.WriteApplication(appToSend)
	c.Equal(ErrInvalidAppStatus, err)
	c.Empty(app)

	appToSend.Status = repository.Orphaned
	appToSend.PayPlanType = "wrong"

	app, err = driver.WriteApplication(appToSend)
	c.Equal(ErrInvalidPayPlanType, err)
	c.Empty(app)
}

func TestPostgresDriver_UpdateApplication(t *testing.T) {
	c := require.New(t)

	db, mock, err := sqlmock.New()
	c.NoError(err)

	defer db.Close()

	driver := NewPostgresDriverFromSQLDBInstance(db)

	mock.ExpectBegin()

	mock.ExpectExec("UPDATE applications").WithArgs("pablo", "ORPHANED", "PAY_AS_YOU_GO_V0", sqlmock.AnyArg(), "60e85042bf95f5003559b791").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery("^SELECT (.+) FROM gateway_settings (.+)").WillReturnRows(sqlmock.NewRows([]string{"application_id", "whitelist_contracts", "whitelist_methods"}).
		AddRow("5f62b7d8be3591c4dea85661",
			`[{"blockchainID":"0021","contracts":["0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48"]}]`,
			`[{"blockchainID":"000C","methods":["\t  eth_getBlockByHash"]}]`))

	mock.ExpectExec("UPDATE gateway_settings").WithArgs("54y4p93body6qco2nrhonz6bltn1k5e8",
		true, `[{"blockchainID":"0021","contracts":["ajua"]}]`, `[{"BlockchainID":"0021","methods":["POST"]}]`,
		pq.StringArray([]string{"url.com"}), pq.StringArray([]string{"gecko.com"}), pq.StringArray([]string{"0021"}), "60e85042bf95f5003559b791").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery("^SELECT (.+) FROM notification_settings (.+)").WillReturnRows(sqlmock.NewRows([]string{"application_id", "signed_up", "on_quarter", "on_half", "on_three_quarters", "on_full"}).
		AddRow("5f62b7d8be3591c4dea85661", false, false, false, false, false))

	mock.ExpectExec("UPDATE notification_settings").WithArgs(true, true, true, true, true, "60e85042bf95f5003559b791").WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	settingsToSend := &repository.GatewaySettings{
		SecretKey:         "54y4p93body6qco2nrhonz6bltn1k5e8",
		SecretKeyRequired: true,
		WhitelistContracts: []repository.WhitelistContract{
			{
				BlockchainID: "0021",
				Contracts:    []string{"ajua"},
			},
		},
		WhitelistMethods: []repository.WhitelistMethod{
			{
				BlockchainID: "0021",
				Methods:      []string{"POST"},
			},
		},
		WhitelistOrigins:     []string{"url.com"},
		WhitelistUserAgents:  []string{"gecko.com"},
		WhitelistBlockchains: []string{"0021"},
	}

	notificationToSend := &repository.NotificationSettings{
		SignedUp:      true,
		Quarter:       true,
		Half:          true,
		ThreeQuarters: true,
		Full:          true,
	}

	err = driver.UpdateApplication("60e85042bf95f5003559b791", &repository.UpdateApplication{
		Name:                 "pablo",
		Status:               repository.Orphaned,
		PayPlanType:          repository.PayAsYouGoV0,
		GatewaySettings:      settingsToSend,
		NotificationSettings: notificationToSend,
	})
	c.NoError(err)

	mock.ExpectBegin()

	mock.ExpectExec("UPDATE applications").WithArgs("pablo", "ORPHANED", "PAY_AS_YOU_GO_V0", sqlmock.AnyArg(), "60e85042bf95f5003559b791").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery("^SELECT (.+) FROM gateway_settings (.+)").WillReturnRows(sqlmock.NewRows(nil))

	mock.ExpectExec("INSERT into gateway_settings").WithArgs("60e85042bf95f5003559b791", "54y4p93body6qco2nrhonz6bltn1k5e8",
		true, `[{"blockchainID":"0021","contracts":["ajua"]}]`, `[{"BlockchainID":"0021","methods":["POST"]}]`,
		pq.StringArray([]string{"url.com"}), pq.StringArray([]string{"gecko.com"}), pq.StringArray([]string{"0021"})).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err = driver.UpdateApplication("60e85042bf95f5003559b791", &repository.UpdateApplication{
		Name:            "pablo",
		Status:          repository.Orphaned,
		PayPlanType:     repository.PayAsYouGoV0,
		GatewaySettings: settingsToSend,
	})
	c.NoError(err)

	mock.ExpectBegin()

	mock.ExpectExec("UPDATE applications").WithArgs("pablo", "ORPHANED", "PAY_AS_YOU_GO_V0", sqlmock.AnyArg(), "60e85042bf95f5003559b791").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery("^SELECT (.+) FROM notification_settings (.+)").WillReturnRows(sqlmock.NewRows(nil))

	mock.ExpectExec("INSERT into notification_settings").WithArgs("60e85042bf95f5003559b791", true, true, true, true, true).WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err = driver.UpdateApplication("60e85042bf95f5003559b791", &repository.UpdateApplication{
		Name:                 "pablo",
		Status:               repository.Orphaned,
		PayPlanType:          repository.PayAsYouGoV0,
		NotificationSettings: notificationToSend,
	})
	c.NoError(err)

	mock.ExpectBegin()

	mock.ExpectExec("UPDATE applications").WithArgs("pablo", "ORPHANED", "PAY_AS_YOU_GO_V0", sqlmock.AnyArg(), "60e85042bf95f5003559b791").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err = driver.UpdateApplication("60e85042bf95f5003559b791", &repository.UpdateApplication{
		Name:        "pablo",
		Status:      repository.Orphaned,
		PayPlanType: repository.PayAsYouGoV0,
	})
	c.NoError(err)

	mock.ExpectBegin()

	mock.ExpectExec("UPDATE applications").WithArgs("pablo", "ORPHANED", "PAY_AS_YOU_GO_V0", sqlmock.AnyArg(), "60e85042bf95f5003559b791").
		WillReturnError(errors.New("error in applications"))

	err = driver.UpdateApplication("60e85042bf95f5003559b791", &repository.UpdateApplication{
		Name:            "pablo",
		Status:          repository.Orphaned,
		PayPlanType:     repository.PayAsYouGoV0,
		GatewaySettings: settingsToSend,
	})
	c.EqualError(err, "error in applications")

	mock.ExpectBegin()

	mock.ExpectExec("UPDATE applications").WithArgs("pablo", "ORPHANED", "PAY_AS_YOU_GO_V0", sqlmock.AnyArg(), "60e85042bf95f5003559b791").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery("^SELECT (.+) FROM gateway_settings (.+)").WillReturnError(errors.New("error reading gateway_settings"))

	err = driver.UpdateApplication("60e85042bf95f5003559b791", &repository.UpdateApplication{
		Name:            "pablo",
		Status:          repository.Orphaned,
		PayPlanType:     repository.PayAsYouGoV0,
		GatewaySettings: settingsToSend,
	})
	c.EqualError(err, "error reading gateway_settings")

	mock.ExpectBegin()

	mock.ExpectExec("UPDATE applications").WithArgs("pablo", "ORPHANED", "PAY_AS_YOU_GO_V0", sqlmock.AnyArg(), "60e85042bf95f5003559b791").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery("^SELECT (.+) FROM gateway_settings (.+)").WillReturnRows(sqlmock.NewRows([]string{"application_id", "whitelist_contracts", "whitelist_methods"}).
		AddRow("5f62b7d8be3591c4dea85661",
			`[{"blockchainID":"0021","contracts":["0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48"]}]`,
			`[{"blockchainID":"000C","methods":["\t  eth_getBlockByHash"]}]`))

	mock.ExpectExec("UPDATE gateway_settings").WithArgs("54y4p93body6qco2nrhonz6bltn1k5e8",
		true, `[{"blockchainID":"0021","contracts":["ajua"]}]`, `[{"BlockchainID":"0021","methods":["POST"]}]`,
		pq.StringArray([]string{"url.com"}), pq.StringArray([]string{"gecko.com"}), pq.StringArray([]string{"0021"}), "60e85042bf95f5003559b791").
		WillReturnError(errors.New("error in settings"))

	err = driver.UpdateApplication("60e85042bf95f5003559b791", &repository.UpdateApplication{
		Name:            "pablo",
		Status:          repository.Orphaned,
		PayPlanType:     repository.PayAsYouGoV0,
		GatewaySettings: settingsToSend,
	})
	c.EqualError(err, "error in settings")

	mock.ExpectBegin()

	mock.ExpectExec("UPDATE applications").WithArgs("pablo", "ORPHANED", "PAY_AS_YOU_GO_V0", sqlmock.AnyArg(), "60e85042bf95f5003559b791").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery("^SELECT (.+) FROM gateway_settings (.+)").WillReturnRows(sqlmock.NewRows(nil))

	mock.ExpectExec("INSERT into gateway_settings").WithArgs("60e85042bf95f5003559b791", "54y4p93body6qco2nrhonz6bltn1k5e8",
		true, `[{"blockchainID":"0021","contracts":["ajua"]}]`, `[{"BlockchainID":"0021","methods":["POST"]}]`,
		pq.StringArray([]string{"url.com"}), pq.StringArray([]string{"gecko.com"}), pq.StringArray([]string{"0021"})).
		WillReturnError(errors.New("error in inserting settings"))

	err = driver.UpdateApplication("60e85042bf95f5003559b791", &repository.UpdateApplication{
		Name:            "pablo",
		Status:          repository.Orphaned,
		PayPlanType:     repository.PayAsYouGoV0,
		GatewaySettings: settingsToSend,
	})
	c.EqualError(err, "error in inserting settings")

	err = driver.UpdateApplication("60e85042bf95f5003559b791", nil)
	c.Equal(ErrNoFieldsToUpdate, err)

	err = driver.UpdateApplication("", nil)
	c.Equal(ErrMissingID, err)

	err = driver.UpdateApplication("60e85042bf95f5003559b791", &repository.UpdateApplication{
		Status: "wrong",
	})
	c.Equal(ErrInvalidAppStatus, err)

	err = driver.UpdateApplication("60e85042bf95f5003559b791", &repository.UpdateApplication{
		Status:      repository.Orphaned,
		PayPlanType: "wrong",
	})
	c.Equal(ErrInvalidPayPlanType, err)
}

func TestPostgresDriver_RemoveApplication(t *testing.T) {
	c := require.New(t)

	db, mock, err := sqlmock.New()
	c.NoError(err)

	defer db.Close()

	driver := NewPostgresDriverFromSQLDBInstance(db)

	mock.ExpectExec("UPDATE applications").WithArgs("AWAITING_GRACE_PERIOD", sqlmock.AnyArg(), "60ddc61b6e2936fhtrns63h2").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = driver.RemoveApplication("60ddc61b6e2936fhtrns63h2")
	c.NoError(err)

	mock.ExpectExec("UPDATE applications").WithArgs("AWAITING_GRACE_PERIOD", sqlmock.AnyArg(), "not-an-id").
		WillReturnError(errors.New("dummy error"))

	err = driver.RemoveApplication("not-an-id")
	c.EqualError(err, "dummy error")

	err = driver.RemoveApplication("")
	c.Equal(ErrMissingID, err)
}
