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

	mock.ExpectQuery("^WITH (.+) SELECT (.+) FROM applications(.+)").WillReturnRows(rows)

	driver := NewPostgresDriverFromSQLDBInstance(db, &ListenerMock{})

	applications, err := driver.ReadApplications()
	c.NoError(err)
	c.Len(applications, 1)

	mock.ExpectQuery("^WITH (.+) SELECT (.+) FROM applications(.+)").WillReturnError(errors.New("dummy error"))

	applications, err = driver.ReadApplications()
	c.EqualError(err, "dummy error")
	c.Empty(applications)
}

func TestPostgresDriver_WriteApplication(t *testing.T) {
	c := require.New(t)

	db, mock, err := sqlmock.New()
	c.NoError(err)

	defer db.Close()

	driver := NewPostgresDriverFromSQLDBInstance(db, &ListenerMock{})

	mock.ExpectBegin()

	mock.ExpectExec("INSERT into applications").WithArgs(sqlmock.AnyArg(),
		"60ddc61b6e29c3003378361D", "klk", "yes@yes.com", "a life", "juancito", "app.com", "ORPHANED", false, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("INSERT into app_limits").WithArgs(sqlmock.AnyArg(), "ENTERPRISE", 2000000).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("INSERT into gateway_aat").WithArgs(sqlmock.AnyArg(),
		"f463b4dd88d865c22acbf38981b6c505bcc46c64",
		"8aaedb01a840fd6c9ab5019786c485bd98e69ca492cdb685aabee8473e7fad77",
		sqlmock.AnyArg(),
		"7a80a331c20cb0ac9e30a0d0c68df5f334b9c8bbe10dcfd95b6cb42bf412037d",
		"1566702d9a667c6007639eeb47a48cd2fed79592c5db9040eadc89f81748a4adef82711854b32065ddafa86eec1f1ed3b6f4f03a0786d9cf12c5262b948d9c01",
		"1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("INSERT into gateway_settings").WithArgs(sqlmock.AnyArg(), "54y4p93body6qco2nrhonz6bltn1k5e8", true, pq.StringArray([]string{"url.com"}), pq.StringArray([]string{"gecko.com"}), pq.StringArray([]string{"0021"})).
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
		Limit: repository.AppLimit{
			PayPlan:     repository.PayPlan{Type: repository.Enterprise},
			CustomLimit: 2000000,
		},
		Status: repository.Orphaned,
		GatewayAAT: repository.GatewayAAT{
			ApplicationPublicKey: "7a80a331c20cb0ac9e30a0d0c68df5f334b9c8bbe10dcfd95b6cb42bf412037d",
			PrivateKey:           "",
			Address:              "f463b4dd88d865c22acbf38981b6c505bcc46c64",
			ApplicationSignature: "1566702d9a667c6007639eeb47a48cd2fed79592c5db9040eadc89f81748a4adef82711854b32065ddafa86eec1f1ed3b6f4f03a0786d9cf12c5262b948d9c01",
			ClientPublicKey:      "8aaedb01a840fd6c9ab5019786c485bd98e69ca492cdb685aabee8473e7fad77",
			Version:              "1",
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
		"60ddc61b6e29c3003378361D", "klk", "yes@yes.com", "a life", "juancito", "app.com", "ORPHANED", false, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(errors.New("error in applications"))

	app, err = driver.WriteApplication(appToSend)
	c.EqualError(err, "error in applications")
	c.Empty(app)

	mock.ExpectBegin()

	mock.ExpectExec("INSERT into applications").WithArgs(sqlmock.AnyArg(),
		"60ddc61b6e29c3003378361D", "klk", "yes@yes.com", "a life", "juancito", "app.com", "ORPHANED", false, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("INSERT into app_limits").WithArgs(sqlmock.AnyArg(), "ENTERPRISE", 2000000).
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
	c.Equal(repository.ErrInvalidAppStatus, err)
	c.Empty(app)

	appToSend.Status = repository.Orphaned
	appToSend.Limit.PayPlan.Type = "wrong"

	app, err = driver.WriteApplication(appToSend)
	c.Equal(repository.ErrInvalidPayPlanType, err)
	c.Empty(app)
}

func TestPostgresDriver_UpdateApplication(t *testing.T) {
	c := require.New(t)

	db, mock, err := sqlmock.New()
	c.NoError(err)

	defer db.Close()

	driver := NewPostgresDriverFromSQLDBInstance(db, &ListenerMock{})

	/* Update works when all fields provided */
	mock.ExpectBegin()

	mock.ExpectExec("UPDATE applications").WithArgs("pablo", "ORPHANED", sqlmock.AnyArg(), sqlmock.AnyArg(), "60e85042bf95f5003559b791").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery("^SELECT (.+) FROM app_limits (.+)").WillReturnRows(sqlmock.NewRows([]string{"application_id", "pay_plan", "custom_limit"}).
		AddRow("5f62b7d8be3591c4dea85661", "PAY_AS_YOU_GO_V0", 0))

	mock.ExpectExec("UPDATE app_limits").WithArgs("PAY_AS_YOU_GO_V0", nil, "60e85042bf95f5003559b791").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery("^SELECT (.+) FROM gateway_settings (.+)").WillReturnRows(sqlmock.NewRows([]string{"application_id"}).
		AddRow("5f62b7d8be3591c4dea85661"))
	mock.ExpectExec("UPDATE gateway_settings").WithArgs("54y4p93body6qco2nrhonz6bltn1k5e8", true, pq.StringArray([]string{"url.com"}), pq.StringArray([]string{"gecko.com"}), pq.StringArray([]string{"0021"}), "60e85042bf95f5003559b791").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery("SELECT (.+) FROM whitelist_contracts (.+)").WillReturnRows(sqlmock.NewRows([]string{"application_id", "blockchain_id"}).
		AddRow("5f62b7d8be3591c4dea85661", "0021"))
	mock.ExpectExec("UPDATE whitelist_contracts").WithArgs(pq.StringArray([]string{"ajua"}), "60e85042bf95f5003559b791", "0021").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery("SELECT (.+) FROM whitelist_methods (.+)").WillReturnRows(sqlmock.NewRows([]string{"application_id", "blockchain_id"}).
		AddRow("5f62b7d8be3591c4dea85661", "0021"))
	mock.ExpectExec("UPDATE whitelist_methods").WithArgs(pq.StringArray([]string{"POST"}), "60e85042bf95f5003559b791", "0021").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery("^SELECT (.+) FROM notification_settings (.+)").WillReturnRows(sqlmock.NewRows([]string{"application_id", "signed_up", "on_quarter", "on_half", "on_three_quarters", "on_full"}).
		AddRow("5f62b7d8be3591c4dea85661", false, false, false, false, false))

	mock.ExpectExec("UPDATE notification_settings").WithArgs(true, true, true, true, true, "60e85042bf95f5003559b791").WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	limitToSend := &repository.AppLimit{
		PayPlan: repository.PayPlan{
			Type:  repository.PayAsYouGoV0,
			Limit: 0,
		},
		CustomLimit: 0,
	}

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
		Limit:                limitToSend,
		GatewaySettings:      settingsToSend,
		NotificationSettings: notificationToSend,
		FirstDateSurpassed:   time.Now(),
	})
	c.NoError(err)

	/* Update works when Notification Settings missing */
	mock.ExpectBegin()

	mock.ExpectExec("UPDATE applications").WithArgs("pablo", "ORPHANED", sqlmock.AnyArg(), sqlmock.AnyArg(), "60e85042bf95f5003559b791").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery("^SELECT (.+) FROM app_limits (.+)").WillReturnRows(sqlmock.NewRows([]string{"application_id", "pay_plan", "custom_limit"}).
		AddRow("5f62b7d8be3591c4dea85661", "PAY_AS_YOU_GO_V0", 0))

	mock.ExpectExec("UPDATE app_limits").WithArgs("PAY_AS_YOU_GO_V0", nil, "60e85042bf95f5003559b791").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery("^SELECT (.+) FROM gateway_settings (.+)").WillReturnRows(sqlmock.NewRows(nil))

	mock.ExpectExec("INSERT into gateway_settings").WithArgs("60e85042bf95f5003559b791", "54y4p93body6qco2nrhonz6bltn1k5e8",
		true, pq.StringArray([]string{"url.com"}), pq.StringArray([]string{"gecko.com"}), pq.StringArray([]string{"0021"})).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery("SELECT (.+) FROM whitelist_contracts (.+)").WillReturnRows(sqlmock.NewRows([]string{"application_id", "blockchain_id"}).
		AddRow("5f62b7d8be3591c4dea85661", "0021"))
	mock.ExpectExec("UPDATE whitelist_contracts").WithArgs(pq.StringArray([]string{"ajua"}), "60e85042bf95f5003559b791", "0021").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery("SELECT (.+) FROM whitelist_methods (.+)").WillReturnRows(sqlmock.NewRows([]string{"application_id", "blockchain_id"}).
		AddRow("5f62b7d8be3591c4dea85661", "0021"))
	mock.ExpectExec("UPDATE whitelist_methods").WithArgs(pq.StringArray([]string{"POST"}), "60e85042bf95f5003559b791", "0021").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err = driver.UpdateApplication("60e85042bf95f5003559b791", &repository.UpdateApplication{
		Name:            "pablo",
		Status:          repository.Orphaned,
		Limit:           limitToSend,
		GatewaySettings: settingsToSend,
	})
	c.NoError(err)

	/* Update works when Gateway Settings and Limit missing */
	mock.ExpectBegin()

	mock.ExpectExec("UPDATE applications").WithArgs("pablo", "ORPHANED", sqlmock.AnyArg(), sqlmock.AnyArg(), "60e85042bf95f5003559b791").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery("^SELECT (.+) FROM notification_settings (.+)").WillReturnRows(sqlmock.NewRows(nil))

	mock.ExpectExec("INSERT into notification_settings").WithArgs("60e85042bf95f5003559b791", true, true, true, true, true).WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err = driver.UpdateApplication("60e85042bf95f5003559b791", &repository.UpdateApplication{
		Name:                 "pablo",
		Status:               repository.Orphaned,
		NotificationSettings: notificationToSend,
	})
	c.NoError(err)

	/* Update works when Notification Settings and Gateway Settings missing */
	mock.ExpectBegin()

	mock.ExpectExec("UPDATE applications").WithArgs("pablo", "ORPHANED", sqlmock.AnyArg(), sqlmock.AnyArg(), "60e85042bf95f5003559b791").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery("^SELECT (.+) FROM app_limits (.+)").WillReturnRows(sqlmock.NewRows([]string{"application_id", "pay_plan", "custom_limit"}).
		AddRow("5f62b7d8be3591c4dea85661", "PAY_AS_YOU_GO_V0", 0))

	mock.ExpectExec("UPDATE app_limits").WithArgs("PAY_AS_YOU_GO_V0", nil, "60e85042bf95f5003559b791").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err = driver.UpdateApplication("60e85042bf95f5003559b791", &repository.UpdateApplication{
		Name:   "pablo",
		Status: repository.Orphaned,
		Limit:  limitToSend,
	})
	c.NoError(err)

	/* Update errors as expected on Applications table Error */
	mock.ExpectBegin()

	mock.ExpectExec("UPDATE applications").WithArgs("pablo", "ORPHANED", sqlmock.AnyArg(), sqlmock.AnyArg(), "60e85042bf95f5003559b791").
		WillReturnError(errors.New("error in applications"))

	err = driver.UpdateApplication("60e85042bf95f5003559b791", &repository.UpdateApplication{
		Name:                 "pablo",
		Status:               repository.Orphaned,
		NotificationSettings: notificationToSend,
		GatewaySettings:      settingsToSend,
	})
	c.EqualError(err, "error in applications")

	/* Update errors as expected on select Error */
	mock.ExpectBegin()

	mock.ExpectExec("UPDATE applications").WithArgs("pablo", "ORPHANED", sqlmock.AnyArg(), sqlmock.AnyArg(), "60e85042bf95f5003559b791").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery("^SELECT (.+) FROM app_limits (.+)").WillReturnRows(sqlmock.NewRows([]string{"application_id", "pay_plan", "custom_limit"}).
		AddRow("5f62b7d8be3591c4dea85661", "PAY_AS_YOU_GO_V0", 0))

	mock.ExpectExec("UPDATE app_limits").WithArgs("PAY_AS_YOU_GO_V0", nil, "60e85042bf95f5003559b791").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery("^SELECT (.+) FROM gateway_settings (.+)").WillReturnError(errors.New("error reading gateway_settings"))

	err = driver.UpdateApplication("60e85042bf95f5003559b791", &repository.UpdateApplication{
		Name:            "pablo",
		Status:          repository.Orphaned,
		Limit:           limitToSend,
		GatewaySettings: settingsToSend,
	})
	c.EqualError(err, "error reading gateway_settings")

	/* Update errors as expected on update Error */
	mock.ExpectBegin()

	mock.ExpectExec("UPDATE applications").WithArgs("pablo", "ORPHANED", sqlmock.AnyArg(), sqlmock.AnyArg(), "60e85042bf95f5003559b791").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery("^SELECT (.+) FROM app_limits (.+)").WillReturnRows(sqlmock.NewRows([]string{"application_id", "pay_plan", "custom_limit"}).
		AddRow("5f62b7d8be3591c4dea85661", "PAY_AS_YOU_GO_V0", 0))

	mock.ExpectExec("UPDATE app_limits").WithArgs("PAY_AS_YOU_GO_V0", nil, "60e85042bf95f5003559b791").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery("^SELECT (.+) FROM gateway_settings (.+)").WillReturnRows(sqlmock.NewRows([]string{"application_id"}).
		AddRow("5f62b7d8be3591c4dea85661"))

	mock.ExpectExec("UPDATE gateway_settings").WithArgs("54y4p93body6qco2nrhonz6bltn1k5e8",
		true, pq.StringArray([]string{"url.com"}), pq.StringArray([]string{"gecko.com"}), pq.StringArray([]string{"0021"}), "60e85042bf95f5003559b791").
		WillReturnError(errors.New("error in settings"))

	err = driver.UpdateApplication("60e85042bf95f5003559b791", &repository.UpdateApplication{
		Name:            "pablo",
		Status:          repository.Orphaned,
		Limit:           limitToSend,
		GatewaySettings: settingsToSend,
	})
	c.EqualError(err, "error in settings")

	/* Update errors as expected on insert Error */
	mock.ExpectBegin()

	mock.ExpectExec("UPDATE applications").WithArgs("pablo", "ORPHANED", sqlmock.AnyArg(), sqlmock.AnyArg(), "60e85042bf95f5003559b791").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery("^SELECT (.+) FROM app_limits (.+)").WillReturnRows(sqlmock.NewRows([]string{"application_id", "pay_plan", "custom_limit"}).
		AddRow("5f62b7d8be3591c4dea85661", "PAY_AS_YOU_GO_V0", 0))

	mock.ExpectExec("UPDATE app_limits").WithArgs("PAY_AS_YOU_GO_V0", nil, "60e85042bf95f5003559b791").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery("^SELECT (.+) FROM gateway_settings (.+)").WillReturnRows(sqlmock.NewRows(nil))

	mock.ExpectExec("INSERT into gateway_settings").WithArgs("60e85042bf95f5003559b791", "54y4p93body6qco2nrhonz6bltn1k5e8",
		true, pq.StringArray([]string{"url.com"}), pq.StringArray([]string{"gecko.com"}), pq.StringArray([]string{"0021"})).
		WillReturnError(errors.New("error in inserting settings"))

	err = driver.UpdateApplication("60e85042bf95f5003559b791", &repository.UpdateApplication{
		Name:            "pablo",
		Status:          repository.Orphaned,
		Limit:           limitToSend,
		GatewaySettings: settingsToSend,
	})
	c.EqualError(err, "error in inserting settings")

	/* Update errors as expected when no fields provided */
	err = driver.UpdateApplication("60e85042bf95f5003559b791", nil)
	c.Equal(ErrNoFieldsToUpdate, err)

	/* Update errors as expected when no ID provided */
	err = driver.UpdateApplication("", nil)
	c.Equal(ErrMissingID, err)

	/* Update errors as expected when invalid status provided */
	err = driver.UpdateApplication("60e85042bf95f5003559b791", &repository.UpdateApplication{
		Status: "wrong",
	})
	c.Equal(repository.ErrInvalidAppStatus, err)

	/* Update errors as expected when invalid pay plan type provided */
	err = driver.UpdateApplication("60e85042bf95f5003559b791", &repository.UpdateApplication{
		Status: repository.Orphaned,
		Limit: &repository.AppLimit{
			PayPlan: repository.PayPlan{
				Type: "wrong",
			},
		},
	})
	c.Equal(repository.ErrInvalidPayPlanType, err)

	/* Update errors as expected when attempting to update a non Enterprise Plan with a custom limit */
	err = driver.UpdateApplication("60e85042bf95f5003559b791", &repository.UpdateApplication{
		Status: repository.Orphaned,
		Limit: &repository.AppLimit{
			PayPlan: repository.PayPlan{
				Type: repository.FreetierV0,
			},
			CustomLimit: 12345,
		},
	})
	c.Equal(repository.ErrNotEnterprisePlan, err)

	/* Update errors as expected when attempting to update an Enterprise Plan without a custom limit */
	err = driver.UpdateApplication("60e85042bf95f5003559b791", &repository.UpdateApplication{
		Status: repository.Orphaned,
		Limit: &repository.AppLimit{
			PayPlan: repository.PayPlan{
				Type: repository.Enterprise,
			},
		},
	})
	c.Equal(repository.ErrEnterprisePlanNeedsCustomLimit, err)
}

func TestPostgresDriver_RemoveApplication(t *testing.T) {
	c := require.New(t)

	db, mock, err := sqlmock.New()
	c.NoError(err)

	defer db.Close()

	driver := NewPostgresDriverFromSQLDBInstance(db, &ListenerMock{})

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

func TestPostgresDriver_UpdateFirstDateSurpassed(t *testing.T) {
	c := require.New(t)

	db, mock, err := sqlmock.New()
	c.NoError(err)

	defer db.Close()

	driver := NewPostgresDriverFromSQLDBInstance(db, &ListenerMock{})

	mock.ExpectExec("UPDATE applications").WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "60ddc61b6e2936fhtrns63h2", "60ddc61b6e2936fhtrns63h3").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = driver.UpdateFirstDateSurpassed(&repository.UpdateFirstDateSurpassed{
		FirstDateSurpassed: time.Now(),
		ApplicationIDs:     []string{"60ddc61b6e2936fhtrns63h2", "60ddc61b6e2936fhtrns63h3"},
	})
	c.NoError(err)

	mock.ExpectExec("UPDATE applications").WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "60ddc61b6e2936fhtrns63h2", "60ddc61b6e2936fhtrns63h3").
		WillReturnError(errors.New("dummy error"))

	err = driver.UpdateFirstDateSurpassed(&repository.UpdateFirstDateSurpassed{
		FirstDateSurpassed: time.Now(),
		ApplicationIDs:     []string{"60ddc61b6e2936fhtrns63h2", "60ddc61b6e2936fhtrns63h3"},
	})
	c.EqualError(err, "dummy error")
}
