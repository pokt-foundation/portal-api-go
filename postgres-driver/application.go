package postgresdriver

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pokt-foundation/portal-api-go/repository"
)

const (
	selectApplications = `
	SELECT a.application_id, a.contact_email, a.created_at, a.description, a.dummy, a.name, a.owner, a.status, a.updated_at, a.url, a.user_id, a.pay_plan_type,
	ga.address AS ga_address, ga.client_public_key AS ga_client_public_key, ga.private_key AS ga_private_key, ga.public_key AS ga_public_key, ga.signature AS ga_signature, ga.version AS ga_version,
	pa.public_key AS pa_public_key, pa.address AS pa_address,
	gs.secret_key, gs.secret_key_required, gs.whitelist_blockchains, gs.whitelist_contracts, gs.whitelist_methods, gs.whitelist_origins, gs.whitelist_user_agents,
	ns.signed_up, ns.on_quarter, ns.on_half, ns.on_three_quarters, ns.on_full
	FROM applications AS a
	LEFT JOIN gateway_aat AS ga ON a.application_id=ga.application_id
	LEFT JOIN public_pocket_account AS pa ON a.application_id=pa.application_id
	LEFT JOIN gateway_settings AS gs ON a.application_id=gs.application_id
	LEFT JOIN notification_settings AS ns ON a.application_id=ns.application_id`
	selectGatewaySettings = `
	SELECT application_id, secret_key, secret_key_required, whitelist_blockchains, whitelist_contracts, whitelist_methods, whitelist_origins, whitelist_user_agents
	FROM gateway_settings WHERE application_id = $1`
	selectNotificationSettings = `
	SELECT application_id, signed_up, on_quarter, on_half, on_three_quarters, on_full
	FROM notification_settings WHERE application_id = $1`
	insertApplicationScript = `
	INSERT into applications (application_id, user_id, name, contact_email, description, owner, url, pay_plan_type, status, created_at, updated_at)
	VALUES (:application_id, :user_id, :name, :contact_email, :description, :owner, :url, :pay_plan_type, :status, :created_at, :updated_at)`
	insertGatewayAATScript = `
	INSERT into gateway_aat (application_id, address, client_public_key, private_key, public_key, signature, version)
	VALUES (:application_id, :address, :client_public_key, :private_key, :public_key, :signature, :version)`
	insertPublicPocketAccountScript = `
	INSERT into public_pocket_account (application_id, public_key, address)
	VALUES (:application_id, :public_key, :address)`
	insertGatewaySettingsScript = `
	INSERT into gateway_settings (application_id, secret_key, secret_key_required, whitelist_contracts, whitelist_methods, whitelist_origins, whitelist_user_agents, whitelist_blockchains)
	VALUES (:application_id, :secret_key, :secret_key_required, :whitelist_contracts, :whitelist_methods, :whitelist_origins, :whitelist_user_agents, :whitelist_blockchains)`
	insertNotificationSettingsScript = `
	INSERT into notification_settings (application_id, signed_up, on_quarter, on_half, on_three_quarters, on_full)
	VALUES (:application_id, :signed_up, :on_quarter, :on_half, :on_three_quarters, :on_full)`
	updateApplication = `
	UPDATE applications
	SET name = COALESCE($1, name), status = COALESCE($2, status), pay_plan_type = COALESCE($3, pay_plan_type), updated_at = $4
	WHERE application_id = $5`
	removeApplication = `
	UPDATE applications
	SET status = COALESCE($1, status), updated_at = $2
	WHERE application_id = $3`
	updateGatewaySettings = `
	UPDATE gateway_settings
	SET secret_key = :secret_key, secret_key_required = :secret_key_required, whitelist_contracts = :whitelist_contracts, whitelist_methods = :whitelist_methods, whitelist_origins = :whitelist_origins, whitelist_user_agents = :whitelist_user_agents, whitelist_blockchains = :whitelist_blockchains
	WHERE application_id = :application_id`
	updateNotificationSettings = `
	UPDATE notification_settings
	SET signed_up = :signed_up, on_quarter = :on_quarter, on_half = :on_half, on_three_quarters = :on_three_quarters, on_full = :on_full
	WHERE application_id = :application_id`
)

var (
	ErrInvalidAppStatus   = errors.New("invalid app status")
	ErrInvalidPayPlanType = errors.New("invalid pay plan type")
)

type dbApplication struct {
	ApplicationID        string         `db:"application_id"`
	UserID               sql.NullString `db:"user_id"`
	Name                 sql.NullString `db:"name"`
	Status               sql.NullString `db:"status"`
	ContactEmail         sql.NullString `db:"contact_email"`
	Description          sql.NullString `db:"description"`
	GAAddress            sql.NullString `db:"ga_address"`
	GAClientPublicKey    sql.NullString `db:"ga_client_public_key"`
	GAPrivateKey         sql.NullString `db:"ga_private_key"`
	GAPublicKey          sql.NullString `db:"ga_public_key"`
	GASignature          sql.NullString `db:"ga_signature"`
	GAVersion            sql.NullString `db:"ga_version"`
	Owner                sql.NullString `db:"owner"`
	PAPublicKey          sql.NullString `db:"pa_public_key"`
	PAAdress             sql.NullString `db:"pa_address"`
	SecretKey            sql.NullString `db:"secret_key"`
	URL                  sql.NullString `db:"url"`
	PayPlanType          sql.NullString `db:"pay_plan_type"`
	WhitelistContracts   sql.NullString `db:"whitelist_contracts"`
	WhitelistMethods     sql.NullString `db:"whitelist_methods"`
	WhitelistOrigins     pq.StringArray `db:"whitelist_origins"`
	WhitelistUserAgents  pq.StringArray `db:"whitelist_user_agents"`
	WhitelistBlockchains pq.StringArray `db:"whitelist_blockchains"`
	Dummy                sql.NullBool   `db:"dummy"`
	SecretKeyRequired    sql.NullBool   `db:"secret_key_required"`
	SignedUp             sql.NullBool   `db:"signed_up"`
	Quarter              sql.NullBool   `db:"on_quarter"`
	Half                 sql.NullBool   `db:"on_half"`
	ThreeQuarters        sql.NullBool   `db:"on_three_quarters"`
	Full                 sql.NullBool   `db:"on_full"`
	CreatedAt            sql.NullTime   `db:"created_at"`
	UpdatedAt            sql.NullTime   `db:"updated_at"`
}

func (a *dbApplication) toApplication() *repository.Application {
	return &repository.Application{
		ID:           a.ApplicationID,
		UserID:       a.UserID.String,
		Name:         a.Name.String,
		Status:       repository.AppStatus(a.Status.String),
		ContactEmail: a.ContactEmail.String,
		Description:  a.Description.String,
		Owner:        a.Owner.String,
		URL:          a.URL.String,
		PayPlanType:  repository.PayPlanType(a.PayPlanType.String),
		Dummy:        a.Dummy.Bool,
		CreatedAt:    a.CreatedAt.Time,
		UpdatedAt:    a.UpdatedAt.Time,
		GatewayAAT: repository.GatewayAAT{
			Address:              a.GAAddress.String,
			ApplicationPublicKey: a.GAPublicKey.String,
			ApplicationSignature: a.GASignature.String,
			ClientPublicKey:      a.GAClientPublicKey.String,
			PrivateKey:           a.GAPrivateKey.String,
			Version:              a.GAVersion.String,
		},
		GatewaySettings: repository.GatewaySettings{
			SecretKey:            a.SecretKey.String,
			SecretKeyRequired:    a.SecretKeyRequired.Bool,
			WhitelistBlockchains: a.WhitelistBlockchains,
			WhitelistContracts:   stringToWhitelistContracts(a.WhitelistContracts),
			WhitelistMethods:     stringToWhitelistMethods(a.WhitelistMethods),
			WhitelistOrigins:     a.WhitelistOrigins,
			WhitelistUserAgents:  a.WhitelistUserAgents,
		},
		NotificationSettings: repository.NotificationSettings{
			SignedUp:      a.SignedUp.Bool,
			Quarter:       a.Quarter.Bool,
			Half:          a.Half.Bool,
			ThreeQuarters: a.ThreeQuarters.Bool,
			Full:          a.Full.Bool,
		},
		PublicPocketAccount: repository.PublicPocketAccount{
			PublicKey: a.PAPublicKey.String,
			Address:   a.PAAdress.String,
		},
	}
}

type insertDBApp struct {
	ApplicationID string         `db:"application_id"`
	UserID        sql.NullString `db:"user_id"`
	Name          sql.NullString `db:"name"`
	ContactEmail  sql.NullString `db:"contact_email"`
	Description   sql.NullString `db:"description"`
	Owner         sql.NullString `db:"owner"`
	URL           sql.NullString `db:"url"`
	PayPlanType   sql.NullString `db:"pay_plan_type"`
	Status        sql.NullString `db:"status"`
	CreatedAt     time.Time      `db:"created_at"`
	UpdatedAt     time.Time      `db:"updated_at"`
}

func extractInsertDBApp(app *repository.Application) *insertDBApp {
	return &insertDBApp{
		ApplicationID: app.ID,
		UserID:        newSQLNullString(app.UserID),
		Name:          newSQLNullString(app.Name),
		ContactEmail:  newSQLNullString(app.ContactEmail),
		Description:   newSQLNullString(app.Description),
		Owner:         newSQLNullString(app.Owner),
		URL:           newSQLNullString(app.URL),
		PayPlanType:   newSQLNullString(string(app.PayPlanType)),
		Status:        newSQLNullString(string(app.Status)),
		CreatedAt:     app.CreatedAt,
		UpdatedAt:     app.UpdatedAt,
	}
}

type insertGatewayAAT struct {
	ApplicationID   string         `db:"application_id"`
	Address         sql.NullString `db:"address"`
	ClientPublicKey sql.NullString `db:"client_public_key"`
	PrivateKey      sql.NullString `db:"private_key"`
	PublicKey       sql.NullString `db:"public_key"`
	Signature       sql.NullString `db:"signature"`
	Version         sql.NullString `db:"version"`
}

func (i *insertGatewayAAT) isNotNull() bool {
	return i.PublicKey.Valid || i.Signature.Valid || i.ClientPublicKey.Valid || i.Version.Valid
}

func extractInsertGatewayAAT(app *repository.Application) *insertGatewayAAT {
	return &insertGatewayAAT{
		ApplicationID:   app.ID,
		Address:         newSQLNullString(app.GatewayAAT.Address),
		ClientPublicKey: newSQLNullString(app.GatewayAAT.ClientPublicKey),
		PrivateKey:      newSQLNullString(app.GatewayAAT.PrivateKey),
		PublicKey:       newSQLNullString(app.GatewayAAT.ApplicationPublicKey),
		Signature:       newSQLNullString(app.GatewayAAT.ApplicationSignature),
		Version:         newSQLNullString(app.GatewayAAT.Version),
	}
}

type insertPublicPocketAccount struct {
	ApplicationID string         `db:"application_id"`
	PublicKey     sql.NullString `db:"public_key"`
	Address       sql.NullString `db:"address"`
}

func (i *insertPublicPocketAccount) isNotNull() bool {
	return i.PublicKey.Valid || i.Address.Valid
}

func extractInsertPublicPocketAccount(app *repository.Application) *insertPublicPocketAccount {
	return &insertPublicPocketAccount{
		ApplicationID: app.ID,
		PublicKey:     newSQLNullString(app.PublicPocketAccount.PublicKey),
		Address:       newSQLNullString(app.PublicPocketAccount.Address),
	}
}

type insertGatewaySettings struct {
	ApplicationID        string         `db:"application_id"`
	SecretKey            sql.NullString `db:"secret_key"`
	SecretKeyRequired    bool           `db:"secret_key_required"`
	WhitelistContracts   sql.NullString `db:"whitelist_contracts"`
	WhitelistMethods     sql.NullString `db:"whitelist_methods"`
	WhitelistOrigins     pq.StringArray `db:"whitelist_origins"`
	WhitelistUserAgents  pq.StringArray `db:"whitelist_user_agents"`
	WhitelistBlockchains pq.StringArray `db:"whitelist_blockchains"`
}

func (i *insertGatewaySettings) isNotNull() bool {
	return i.SecretKey.Valid || i.WhitelistContracts.Valid || i.WhitelistMethods.Valid ||
		len(i.WhitelistOrigins) != 0 || len(i.WhitelistUserAgents) != 0 || len(i.WhitelistBlockchains) != 0
}

func (i *insertGatewaySettings) isUpdatable() bool {
	return i != nil
}

func (i *insertGatewaySettings) read(appID string, driver *PostgresDriver) (updatable, error) {
	var settings insertGatewaySettings

	err := driver.Get(&settings, selectGatewaySettings, appID)
	if err != nil {
		return nil, err
	}

	return &settings, nil
}

func marshalWhitelistContractsAndMethods(contracts []repository.WhitelistContract, methods []repository.WhitelistMethod) (string, string) {
	var marshaledWhitelistContracts []byte
	if len(contracts) > 0 {
		marshaledWhitelistContracts, _ = json.Marshal(contracts)
	}

	var marshaledWhitelistMethods []byte
	if len(methods) > 0 {
		marshaledWhitelistMethods, _ = json.Marshal(methods)
	}

	return string(marshaledWhitelistContracts), string(marshaledWhitelistMethods)
}

func convertRepositoryToDBGatewaySettings(id string, settings *repository.GatewaySettings) *insertGatewaySettings {
	if settings == nil {
		return nil
	}

	marshaledWhitelistContracts, marshaledWhitelistMethods := marshalWhitelistContractsAndMethods(settings.WhitelistContracts,
		settings.WhitelistMethods)

	return &insertGatewaySettings{
		ApplicationID:        id,
		SecretKey:            newSQLNullString(settings.SecretKey),
		SecretKeyRequired:    settings.SecretKeyRequired,
		WhitelistContracts:   newSQLNullString(marshaledWhitelistContracts),
		WhitelistMethods:     newSQLNullString(marshaledWhitelistMethods),
		WhitelistOrigins:     settings.WhitelistOrigins,
		WhitelistUserAgents:  settings.WhitelistUserAgents,
		WhitelistBlockchains: settings.WhitelistBlockchains,
	}
}

func extractInsertGatewaySettings(app *repository.Application) *insertGatewaySettings {
	marshaledWhitelistContracts, marshaledWhitelistMethods := marshalWhitelistContractsAndMethods(app.GatewaySettings.WhitelistContracts,
		app.GatewaySettings.WhitelistMethods)

	return &insertGatewaySettings{
		ApplicationID:        app.ID,
		SecretKey:            newSQLNullString(app.GatewaySettings.SecretKey),
		SecretKeyRequired:    app.GatewaySettings.SecretKeyRequired,
		WhitelistContracts:   newSQLNullString(marshaledWhitelistContracts),
		WhitelistMethods:     newSQLNullString(marshaledWhitelistMethods),
		WhitelistOrigins:     app.GatewaySettings.WhitelistOrigins,
		WhitelistUserAgents:  app.GatewaySettings.WhitelistUserAgents,
		WhitelistBlockchains: app.GatewaySettings.WhitelistBlockchains,
	}
}

func stringToWhitelistContracts(rawContracts sql.NullString) []repository.WhitelistContract {
	contracts := []repository.WhitelistContract{}

	if !rawContracts.Valid {
		return contracts
	}

	_ = json.Unmarshal([]byte(rawContracts.String), &contracts)

	for i, contract := range contracts {
		for j, inContract := range contract.Contracts {
			contracts[i].Contracts[j] = strings.TrimSpace(inContract)
		}
	}

	return contracts
}

func stringToWhitelistMethods(rawMethods sql.NullString) []repository.WhitelistMethod {
	methods := []repository.WhitelistMethod{}

	if !rawMethods.Valid {
		return methods
	}

	_ = json.Unmarshal([]byte(rawMethods.String), &methods)

	for i, method := range methods {
		for j, inMethod := range method.Methods {
			methods[i].Methods[j] = strings.TrimSpace(inMethod)
		}
	}

	return methods
}

type insertNotificationSettings struct {
	ApplicationID string `db:"application_id"`
	SignedUp      bool   `db:"signed_up"`
	Quarter       bool   `db:"on_quarter"`
	Half          bool   `db:"on_half"`
	ThreeQuarters bool   `db:"on_three_quarters"`
	Full          bool   `db:"on_full"`
}

func (i *insertNotificationSettings) isNotNull() bool {
	return true
}

func (i *insertNotificationSettings) isUpdatable() bool {
	return i != nil
}

func (i *insertNotificationSettings) read(appID string, driver *PostgresDriver) (updatable, error) {
	var settings insertNotificationSettings

	err := driver.Get(&settings, selectNotificationSettings, appID)
	if err != nil {
		return nil, err
	}

	return &settings, nil
}

func extractInsertNotificationSettings(app *repository.Application) *insertNotificationSettings {
	return &insertNotificationSettings{
		ApplicationID: app.ID,
		SignedUp:      app.NotificationSettings.SignedUp,
		Quarter:       app.NotificationSettings.Quarter,
		Half:          app.NotificationSettings.Half,
		ThreeQuarters: app.NotificationSettings.ThreeQuarters,
		Full:          app.NotificationSettings.Full,
	}
}

func convertRepositoryToDBNotificationSettings(id string, settings *repository.NotificationSettings) *insertNotificationSettings {
	if settings == nil {
		return nil
	}

	return &insertNotificationSettings{
		ApplicationID: id,
		SignedUp:      settings.SignedUp,
		Quarter:       settings.Quarter,
		Half:          settings.Half,
		ThreeQuarters: settings.ThreeQuarters,
		Full:          settings.Full,
	}
}

// ReadApplications returns all applications on the database
func (d *PostgresDriver) ReadApplications() ([]*repository.Application, error) {
	var dbApplications []*dbApplication

	err := d.Select(&dbApplications, selectApplications)
	if err != nil {
		return nil, err
	}

	var applications []*repository.Application

	for _, dbApplication := range dbApplications {
		applications = append(applications, dbApplication.toApplication())
	}

	return applications, nil
}

type nullable interface {
	isNotNull() bool
}

// WriteApplication saves input application in the database
func (d *PostgresDriver) WriteApplication(app *repository.Application) (*repository.Application, error) {
	if !repository.ValidAppStatuses[app.Status] {
		return nil, ErrInvalidAppStatus
	}

	if !repository.ValidPayPlanTypes[app.PayPlanType] {
		return nil, ErrInvalidPayPlanType
	}

	id, err := generateRandomID()
	if err != nil {
		return nil, err
	}

	app.ID = id
	app.CreatedAt = time.Now()
	app.UpdatedAt = time.Now()

	insertApp := extractInsertDBApp(app)

	nullables := []nullable{}
	nullablesScripts := []string{}

	nullables = append(nullables, extractInsertGatewayAAT(app))
	nullablesScripts = append(nullablesScripts, insertGatewayAATScript)

	nullables = append(nullables, extractInsertPublicPocketAccount(app))
	nullablesScripts = append(nullablesScripts, insertPublicPocketAccountScript)

	nullables = append(nullables, extractInsertGatewaySettings(app))
	nullablesScripts = append(nullablesScripts, insertGatewaySettingsScript)

	nullables = append(nullables, extractInsertNotificationSettings(app))
	nullablesScripts = append(nullablesScripts, insertNotificationSettingsScript)

	tx, err := d.Beginx()
	if err != nil {
		return nil, err
	}

	_, err = tx.NamedExec(insertApplicationScript, insertApp)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(nullables); i++ {
		if nullables[i].isNotNull() {
			_, err = tx.NamedExec(nullablesScripts[i], nullables[i])
			if err != nil {
				return nil, err
			}
		}
	}

	return app, tx.Commit()
}

type updatable interface {
	isUpdatable() bool
	read(appID string, driver *PostgresDriver) (updatable, error)
}

type update struct {
	insertScript string
	updateScript string
	toUpdate     updatable
}

func (d *PostgresDriver) doUpdate(id string, update *update, tx *sqlx.Tx) error {
	if !update.toUpdate.isUpdatable() {
		return nil
	}

	_, err := update.toUpdate.read(id, d)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			_, err = tx.NamedExec(update.insertScript, update.toUpdate)
			if err != nil {
				return err
			}

			return nil
		}

		return err
	}

	_, err = tx.NamedExec(update.updateScript, update.toUpdate)
	if err != nil {
		return err
	}

	return nil
}

// UpdateApplication updates fields available in options in db
func (d *PostgresDriver) UpdateApplication(id string, fieldsToUpdate *repository.UpdateApplication) error {
	if id == "" {
		return ErrMissingID
	}

	if fieldsToUpdate == nil {
		return ErrNoFieldsToUpdate
	}

	if !repository.ValidAppStatuses[fieldsToUpdate.Status] {
		return ErrInvalidAppStatus
	}

	if !repository.ValidPayPlanTypes[fieldsToUpdate.PayPlanType] {
		return ErrInvalidPayPlanType
	}

	tx, err := d.Beginx()
	if err != nil {
		return err
	}

	_, err = tx.Exec(updateApplication, newSQLNullString(fieldsToUpdate.Name), newSQLNullString(string(fieldsToUpdate.Status)),
		newSQLNullString(string(fieldsToUpdate.PayPlanType)), time.Now(), id)
	if err != nil {
		return err
	}

	updates := []*update{}

	updates = append(updates, &update{
		insertScript: insertGatewaySettingsScript,
		updateScript: updateGatewaySettings,
		toUpdate:     convertRepositoryToDBGatewaySettings(id, fieldsToUpdate.GatewaySettings),
	})

	updates = append(updates, &update{
		insertScript: insertNotificationSettingsScript,
		updateScript: updateNotificationSettings,
		toUpdate:     convertRepositoryToDBNotificationSettings(id, fieldsToUpdate.NotificationSettings),
	})

	for _, update := range updates {
		err = d.doUpdate(id, update, tx)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// RemoveApplication updates fields available in options in db
func (d *PostgresDriver) RemoveApplication(id string) error {
	if id == "" {
		return ErrMissingID
	}

	_, err := d.Exec(removeApplication, newSQLNullString(string(repository.AwaitingGracePeriod)), time.Now(), id)
	if err != nil {
		return err
	}

	return nil
}
