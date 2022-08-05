package postgresdriver

import (
	"database/sql"
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
		"60ddc61b6e29c3003378361D", "klk", "yes@yes.com", "a life", "juancito", "app.com", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("INSERT into freetier_aat").WithArgs(sqlmock.AnyArg(),
		"7a80a331c20cb0ac9e30a0d0c68df5f334b9c8bbe10dcfd95b6cb42bf412037d",
		"1566702d9a667c6007639eeb47a48cd2fed79592c5db9040eadc89f81748a4adef82711854b32065ddafa86eec1f1ed3b6f4f03a0786d9cf12c5262b948d9c01",
		"8aaedb01a840fd6c9ab5019786c485bd98e69ca492cdb685aabee8473e7fad77", "1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("INSERT into gateway_aat").WithArgs(sqlmock.AnyArg(),
		"7a80a331c20cb0ac9e30a0d0c68df5f334b9c8bbe10dcfd95b6cb42bf412037d",
		"1566702d9a667c6007639eeb47a48cd2fed79592c5db9040eadc89f81748a4adef82711854b32065ddafa86eec1f1ed3b6f4f03a0786d9cf12c5262b948d9c01",
		"8aaedb01a840fd6c9ab5019786c485bd98e69ca492cdb685aabee8473e7fad77", "1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("INSERT into freetier_app_account").WithArgs(sqlmock.AnyArg(),
		"7a80a331c20cb0ac9e30a0d0c68df5f334b9c8bbe10dcfd95b6cb42bf412037d",
		"f463b4dd88d865c22acbf38981b6c505bcc46c64", sql.NullString{}, "1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("INSERT into public_pocket_account").WithArgs(sqlmock.AnyArg(),
		"7a80a331c20cb0ac9e30a0d0c68df5f334b9c8bbe10dcfd95b6cb42bf412037d", "f463b4dd88d865c22acbf38981b6c505bcc46c64").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("INSERT into gateway_settings").WithArgs(sqlmock.AnyArg(),
		"54y4p93body6qco2nrhonz6bltn1k5e8", true, `[{"blockchainID":"0021","contracts":["ajua"]}]`,
		`[{"BlockchainID":"0021","methods":["POST"]}]`, pq.StringArray([]string{"url.com"}), pq.StringArray([]string{"gecko.com"})).
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
		FreeTierAAT: repository.FreeTierAAT{
			ApplicationPublicKey: "7a80a331c20cb0ac9e30a0d0c68df5f334b9c8bbe10dcfd95b6cb42bf412037d",
			ApplicationSignature: "1566702d9a667c6007639eeb47a48cd2fed79592c5db9040eadc89f81748a4adef82711854b32065ddafa86eec1f1ed3b6f4f03a0786d9cf12c5262b948d9c01",
			ClientPublicKey:      "8aaedb01a840fd6c9ab5019786c485bd98e69ca492cdb685aabee8473e7fad77",
			Version:              "1",
		},
		GatewayAAT: repository.GatewayAAT{
			ApplicationPublicKey: "7a80a331c20cb0ac9e30a0d0c68df5f334b9c8bbe10dcfd95b6cb42bf412037d",
			ApplicationSignature: "1566702d9a667c6007639eeb47a48cd2fed79592c5db9040eadc89f81748a4adef82711854b32065ddafa86eec1f1ed3b6f4f03a0786d9cf12c5262b948d9c01",
			ClientPublicKey:      "8aaedb01a840fd6c9ab5019786c485bd98e69ca492cdb685aabee8473e7fad77",
			Version:              "1",
		},
		FreeTierApplicationAccount: repository.FreeTierApplicationAccount{
			PublicKey:  "7a80a331c20cb0ac9e30a0d0c68df5f334b9c8bbe10dcfd95b6cb42bf412037d",
			PrivateKey: "",
			Address:    "f463b4dd88d865c22acbf38981b6c505bcc46c64",
			Version:    "1",
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
			WhitelistOrigins:    []string{"url.com"},
			WhitelistUserAgents: []string{"gecko.com"},
		},
	}

	err = driver.WriteApplication(appToSend)
	c.NoError(err)

	mock.ExpectBegin()

	mock.ExpectExec("INSERT into applications").WithArgs(sqlmock.AnyArg(),
		"60ddc61b6e29c3003378361D", "klk", "yes@yes.com", "a life", "juancito", "app.com", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(errors.New("error in applications"))

	err = driver.WriteApplication(appToSend)
	c.EqualError(err, "error in applications")

	mock.ExpectBegin()

	mock.ExpectExec("INSERT into applications").WithArgs(sqlmock.AnyArg(),
		"60ddc61b6e29c3003378361D", "klk", "yes@yes.com", "a life", "juancito", "app.com", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("INSERT into freetier_aat").WithArgs(sqlmock.AnyArg(),
		"7a80a331c20cb0ac9e30a0d0c68df5f334b9c8bbe10dcfd95b6cb42bf412037d",
		"1566702d9a667c6007639eeb47a48cd2fed79592c5db9040eadc89f81748a4adef82711854b32065ddafa86eec1f1ed3b6f4f03a0786d9cf12c5262b948d9c01",
		"8aaedb01a840fd6c9ab5019786c485bd98e69ca492cdb685aabee8473e7fad77", "1").
		WillReturnError(errors.New("error in freetier_aat"))

	err = driver.WriteApplication(appToSend)
	c.EqualError(err, "error in freetier_aat")
}

func TestPostgresDriver_UpdateApplication(t *testing.T) {
	c := require.New(t)

	db, mock, err := sqlmock.New()
	c.NoError(err)

	defer db.Close()

	driver := NewPostgresDriverFromSQLDBInstance(db)

	mock.ExpectBegin()

	mock.ExpectExec("UPDATE applications").WithArgs("pablo", sqlmock.AnyArg(), "60e85042bf95f5003559b791").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("UPDATE gateway_settings").WithArgs("54y4p93body6qco2nrhonz6bltn1k5e8",
		true, `[{"blockchainID":"0021","contracts":["ajua"]}]`, `[{"BlockchainID":"0021","methods":["POST"]}]`,
		pq.StringArray([]string{"url.com"}), pq.StringArray([]string{"gecko.com"}), "60e85042bf95f5003559b791").
		WillReturnResult(sqlmock.NewResult(1, 1))

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
		WhitelistOrigins:    []string{"url.com"},
		WhitelistUserAgents: []string{"gecko.com"},
	}

	err = driver.UpdateApplication("60e85042bf95f5003559b791", &UpdateApplicationOptions{
		Name:            "pablo",
		GatewatSettings: settingsToSend,
	})
	c.NoError(err)

	mock.ExpectBegin()

	mock.ExpectExec("UPDATE applications").WithArgs("pablo", sqlmock.AnyArg(), "60e85042bf95f5003559b791").
		WillReturnError(errors.New("error in applications"))

	err = driver.UpdateApplication("60e85042bf95f5003559b791", &UpdateApplicationOptions{
		Name:            "pablo",
		GatewatSettings: settingsToSend,
	})
	c.EqualError(err, "error in applications")

	mock.ExpectBegin()

	mock.ExpectExec("UPDATE applications").WithArgs("pablo", sqlmock.AnyArg(), "60e85042bf95f5003559b791").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("UPDATE gateway_settings").WithArgs("54y4p93body6qco2nrhonz6bltn1k5e8",
		true, `[{"blockchainID":"0021","contracts":["ajua"]}]`, `[{"BlockchainID":"0021","methods":["POST"]}]`,
		pq.StringArray([]string{"url.com"}), pq.StringArray([]string{"gecko.com"}), "60e85042bf95f5003559b791").
		WillReturnError(errors.New("error in settings"))

	err = driver.UpdateApplication("60e85042bf95f5003559b791", &UpdateApplicationOptions{
		Name:            "pablo",
		GatewatSettings: settingsToSend,
	})
	c.EqualError(err, "error in settings")

	err = driver.UpdateApplication("60e85042bf95f5003559b791", nil)
	c.Equal(ErrNoFieldsToUpdate, err)
}
