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
	SELECT a.application_id, a.contact_email, a.created_at, a.description, a.dummy, a.name, a.owner, a.status, a.updated_at, a.url, a.user_id, a.first_date_surpassed, 
	ga.address AS ga_address, ga.client_public_key AS ga_client_public_key, ga.private_key AS ga_private_key, ga.public_key AS ga_public_key, ga.signature AS ga_signature, ga.version AS ga_version,
	gs.secret_key, gs.secret_key_required, gs.whitelist_blockchains, gs.whitelist_contracts, gs.whitelist_methods, gs.whitelist_origins, gs.whitelist_user_agents,
	ns.signed_up, ns.on_quarter, ns.on_half, ns.on_three_quarters, ns.on_full,
	al.custom_limit, al.pay_plan, pp.daily_limit as plan_limit
	FROM applications AS a
	LEFT JOIN gateway_aat AS ga ON a.application_id=ga.application_id
	LEFT JOIN gateway_settings AS gs ON a.application_id=gs.application_id
	LEFT JOIN notification_settings AS ns ON a.application_id=ns.application_id
	LEFT JOIN app_limits AS al ON a.application_id=al.application_id
	LEFT JOIN pay_plans AS pp ON al.pay_plan=pp.plan_type`
	selectAppLimit = `
	SELECT application_id, pay_plan, custom_limit
	FROM app_limits WHERE application_id = $1`
	selectGatewaySettings = `
	SELECT application_id, secret_key, secret_key_required, whitelist_blockchains, whitelist_contracts, whitelist_methods, whitelist_origins, whitelist_user_agents
	FROM gateway_settings WHERE application_id = $1`
	selectNotificationSettings = `
	SELECT application_id, signed_up, on_quarter, on_half, on_three_quarters, on_full
	FROM notification_settings WHERE application_id = $1`
	insertApplicationScript = `
	INSERT into applications (application_id, user_id, name, contact_email, description, owner, url, status, dummy, created_at, updated_at)
	VALUES (:application_id, :user_id, :name, :contact_email, :description, :owner, :url, :status, :dummy, :created_at, :updated_at)`
	insertAppLimitScript = `
	INSERT into app_limits (application_id, pay_plan, custom_limit)
	VALUES (:application_id, :pay_plan, :custom_limit)`
	insertGatewayAATScript = `
	INSERT into gateway_aat (application_id, address, client_public_key, private_key, public_key, signature, version)
	VALUES (:application_id, :address, :client_public_key, :private_key, :public_key, :signature, :version)`
	insertGatewaySettingsScript = `
	INSERT into gateway_settings (application_id, secret_key, secret_key_required, whitelist_contracts, whitelist_methods, whitelist_origins, whitelist_user_agents, whitelist_blockchains)
	VALUES (:application_id, :secret_key, :secret_key_required, :whitelist_contracts, :whitelist_methods, :whitelist_origins, :whitelist_user_agents, :whitelist_blockchains)`
	insertNotificationSettingsScript = `
	INSERT into notification_settings (application_id, signed_up, on_quarter, on_half, on_three_quarters, on_full)
	VALUES (:application_id, :signed_up, :on_quarter, :on_half, :on_three_quarters, :on_full)`
	updateApplication = `
	UPDATE applications
	SET name = COALESCE($1, name), status = COALESCE($2, status), first_date_surpassed = COALESCE($3, first_date_surpassed), updated_at = $4
	WHERE application_id = $5`
	updateAppLimitScript = `
	UPDATE app_limits
	SET pay_plan = :pay_plan, custom_limit = :custom_limit
	WHERE application_id = :application_id`
	updateFirstDateSurpassedScript = `
	UPDATE applications
	SET first_date_surpassed = :first_date_surpassed, updated_at = :updated_at
	WHERE application_id IN(:application_ids)`
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
	FirstDateSurpassed   sql.NullTime   `db:"first_date_surpassed"`
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
	PlanType             sql.NullString `db:"pay_plan"`
	PlanLimit            sql.NullInt32  `db:"plan_limit"`
	CustomLimit          sql.NullInt32  `db:"custom_limit"`
	CreatedAt            sql.NullTime   `db:"created_at"`
	UpdatedAt            sql.NullTime   `db:"updated_at"`
}

func (a *dbApplication) toApplication() *repository.Application {
	return &repository.Application{
		ID:                 a.ApplicationID,
		UserID:             a.UserID.String,
		Name:               a.Name.String,
		Status:             repository.AppStatus(a.Status.String),
		ContactEmail:       a.ContactEmail.String,
		Description:        a.Description.String,
		Owner:              a.Owner.String,
		URL:                a.URL.String,
		Dummy:              a.Dummy.Bool,
		FirstDateSurpassed: a.FirstDateSurpassed.Time,
		CreatedAt:          a.CreatedAt.Time,
		UpdatedAt:          a.UpdatedAt.Time,

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
			WhitelistContracts:   nullStringToWhitelistContracts(a.WhitelistContracts),
			WhitelistMethods:     nullStringToWhitelistMethods(a.WhitelistMethods),
			WhitelistOrigins:     a.WhitelistOrigins,
			WhitelistUserAgents:  a.WhitelistUserAgents,
		},
		Limit: repository.AppLimit{
			PayPlan: repository.PayPlan{
				Type:  repository.PayPlanType(a.PlanType.String),
				Limit: int(a.PlanLimit.Int32),
			},
			CustomLimit: int(a.CustomLimit.Int32),
		},
		NotificationSettings: repository.NotificationSettings{
			SignedUp:      a.SignedUp.Bool,
			Quarter:       a.Quarter.Bool,
			Half:          a.Half.Bool,
			ThreeQuarters: a.ThreeQuarters.Bool,
			Full:          a.Full.Bool,
		},
	}
}

type dbAppJSON struct {
	ApplicationID      string `json:"application_id"`
	UserID             string `json:"user_id"`
	Name               string `json:"name"`
	ContactEmail       string `json:"contact_email"`
	Description        string `json:"description"`
	Owner              string `json:"owner"`
	URL                string `json:"url"`
	Status             string `json:"status"`
	CreatedAt          string `json:"created_at"`
	UpdatedAt          string `json:"updated_at"`
	FirstDateSurpassed string `json:"first_date_surpassed"`
	Dummy              bool   `json:"dummy"`
}

func (j dbAppJSON) toOutput() *repository.Application {
	return &repository.Application{
		ID:                 j.ApplicationID,
		UserID:             j.UserID,
		Name:               j.Name,
		ContactEmail:       j.ContactEmail,
		Description:        j.Description,
		Owner:              j.Owner,
		URL:                j.URL,
		Status:             repository.AppStatus(j.Status),
		CreatedAt:          psqlDateToTime(j.CreatedAt),
		UpdatedAt:          psqlDateToTime(j.UpdatedAt),
		FirstDateSurpassed: psqlDateToTime(j.FirstDateSurpassed),
		Dummy:              j.Dummy,
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
	Status        sql.NullString `db:"status"`
	CreatedAt     time.Time      `db:"created_at"`
	UpdatedAt     time.Time      `db:"updated_at"`
	Dummy         bool           `db:"dummy"`
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
		Status:        newSQLNullString(string(app.Status)),
		CreatedAt:     app.CreatedAt,
		UpdatedAt:     app.UpdatedAt,
		Dummy:         app.Dummy,
	}
}

type dbAppLimitJSON struct {
	ApplicationID string             `json:"application_id"`
	PayPlan       repository.PayPlan `json:"pay_plan"`
	CustomLimit   int                `json:"custom_limit"`
}

func (j dbAppLimitJSON) toOutput() *repository.AppLimit {
	return &repository.AppLimit{
		ID:          j.ApplicationID,
		PayPlan:     j.PayPlan,
		CustomLimit: j.CustomLimit,
	}
}

type insertAppLimit struct {
	ApplicationID string         `db:"application_id"`
	PayPlan       sql.NullString `db:"pay_plan"`
	CustomLimit   sql.NullInt32  `db:"custom_limit"`
}

func (i *insertAppLimit) isNotNull() bool {
	return i.PayPlan.Valid || i.CustomLimit.Valid
}

func (i *insertAppLimit) isUpdatable() bool {
	return i != nil
}

func (i *insertAppLimit) read(appID string, driver *PostgresDriver) (updatable, error) {
	var limit insertAppLimit

	err := driver.Get(&limit, selectAppLimit, appID)
	if err != nil {
		return nil, err
	}

	return &limit, nil
}

func extractInsertDBAppLimit(app *repository.Application) *insertAppLimit {
	return &insertAppLimit{
		ApplicationID: app.ID,
		PayPlan:       newSQLNullString(string(app.Limit.PayPlan.Type)),
		CustomLimit:   newSQLNullInt32(int32(app.Limit.CustomLimit)),
	}
}

func convertRepositoryToDBAppLimit(id string, limit *repository.AppLimit) *insertAppLimit {
	if limit == nil {
		return nil
	}

	return &insertAppLimit{
		ApplicationID: id,
		PayPlan:       newSQLNullString(string(limit.PayPlan.Type)),
		CustomLimit:   newSQLNullInt32(int32(limit.CustomLimit)),
	}
}

type dbGatewayAATJSON struct {
	ApplicationID   string `json:"application_id"`
	Address         string `json:"address"`
	ClientPublicKey string `json:"client_public_key"`
	PrivateKey      string `json:"private_key"`
	PublicKey       string `json:"public_key"`
	Signature       string `json:"signature"`
	Version         string `json:"version"`
}

func (j dbGatewayAATJSON) toOutput() *repository.GatewayAAT {
	return &repository.GatewayAAT{
		ID:                   j.ApplicationID,
		Address:              j.Address,
		ClientPublicKey:      j.ClientPublicKey,
		PrivateKey:           j.PrivateKey,
		ApplicationPublicKey: j.PublicKey,
		ApplicationSignature: j.Signature,
		Version:              j.Version,
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
	return i.Address.Valid || i.PublicKey.Valid || i.Signature.Valid || i.ClientPublicKey.Valid || i.Version.Valid || i.PrivateKey.Valid
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

type dbGatewaySettingsJSON struct {
	ApplicationID        string   `json:"application_id"`
	SecretKey            string   `json:"secret_key"`
	SecretKeyRequired    bool     `json:"secret_key_required"`
	WhitelistContracts   string   `json:"whitelist_contracts"`
	WhitelistMethods     string   `json:"whitelist_methods"`
	WhitelistOrigins     []string `json:"whitelist_origins"`
	WhitelistUserAgents  []string `json:"whitelist_user_agents"`
	WhitelistBlockchains []string `json:"whitelist_blockchains"`
}

func (j dbGatewaySettingsJSON) toOutput() *repository.GatewaySettings {
	return &repository.GatewaySettings{
		ID:                   j.ApplicationID,
		SecretKey:            j.SecretKey,
		SecretKeyRequired:    j.SecretKeyRequired,
		WhitelistContracts:   stringToWhitelistContracts(j.WhitelistContracts),
		WhitelistMethods:     stringToWhitelistMethods(j.WhitelistMethods),
		WhitelistOrigins:     j.WhitelistOrigins,
		WhitelistUserAgents:  j.WhitelistUserAgents,
		WhitelistBlockchains: j.WhitelistBlockchains,
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

func nullStringToWhitelistContracts(rawContracts sql.NullString) []repository.WhitelistContract {
	if !rawContracts.Valid {
		return nil
	}

	return stringToWhitelistContracts(rawContracts.String)
}

func stringToWhitelistContracts(rawContracts string) []repository.WhitelistContract {
	contracts := []repository.WhitelistContract{}

	_ = json.Unmarshal([]byte(rawContracts), &contracts)

	for i, contract := range contracts {
		for j, inContract := range contract.Contracts {
			contracts[i].Contracts[j] = strings.TrimSpace(inContract)
		}
	}

	return contracts
}

func nullStringToWhitelistMethods(rawMethods sql.NullString) []repository.WhitelistMethod {
	if !rawMethods.Valid {
		return nil
	}

	return stringToWhitelistMethods(rawMethods.String)
}

func stringToWhitelistMethods(rawMethods string) []repository.WhitelistMethod {
	methods := []repository.WhitelistMethod{}

	_ = json.Unmarshal([]byte(rawMethods), &methods)

	for i, method := range methods {
		for j, inMethod := range method.Methods {
			methods[i].Methods[j] = strings.TrimSpace(inMethod)
		}
	}

	return methods
}

type dbNotificationSettingsJSON struct {
	ApplicationID string `json:"application_id"`
	SignedUp      bool   `json:"signed_up"`
	Quarter       bool   `json:"on_quarter"`
	Half          bool   `json:"on_half"`
	ThreeQuarters bool   `json:"on_three_quarters"`
	Full          bool   `json:"on_full"`
}

func (j dbNotificationSettingsJSON) toOutput() *repository.NotificationSettings {
	return &repository.NotificationSettings{
		ID:            j.ApplicationID,
		SignedUp:      j.SignedUp,
		Quarter:       j.Quarter,
		Half:          j.Half,
		ThreeQuarters: j.ThreeQuarters,
		Full:          j.Full,
	}
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

// WriteApplication saves input application in the database
func (d *PostgresDriver) WriteApplication(app *repository.Application) (*repository.Application, error) {
	if !repository.ValidAppStatuses[app.Status] {
		return nil, ErrInvalidAppStatus
	}

	if !repository.ValidPayPlanTypes[app.Limit.PayPlan.Type] {
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
	insertAppLimit := extractInsertDBAppLimit(app)

	nullables := []nullable{}
	nullablesScripts := []string{}

	nullables = append(nullables, extractInsertGatewayAAT(app))
	nullablesScripts = append(nullablesScripts, insertGatewayAATScript)

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

	_, err = tx.NamedExec(insertAppLimitScript, insertAppLimit)
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

	if fieldsToUpdate.Limit != nil && !repository.ValidPayPlanTypes[fieldsToUpdate.Limit.PayPlan.Type] {
		return ErrInvalidPayPlanType
	}

	tx, err := d.Beginx()
	if err != nil {
		return err
	}

	_, err = tx.Exec(updateApplication, newSQLNullString(fieldsToUpdate.Name), newSQLNullString(string(fieldsToUpdate.Status)),
		newSQLNullTime(fieldsToUpdate.FirstDateSurpassed), time.Now(), id)
	if err != nil {
		return err
	}

	updates := []*update{}

	updates = append(updates, &update{
		insertScript: insertAppLimitScript,
		updateScript: updateAppLimitScript,
		toUpdate:     convertRepositoryToDBAppLimit(id, fieldsToUpdate.Limit),
	})

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

type updateFirstDateSurpassed struct {
	ApplicationIDs     []string  `db:"application_ids"`
	FirstDateSurpassed time.Time `db:"first_date_surpassed"`
	UpdatedAt          time.Time `db:"updated_at"`
}

func (d *PostgresDriver) UpdateFirstDateSurpassed(firstDateSurpassed *repository.UpdateFirstDateSurpassed) error {
	query, args, err := sqlx.Named(updateFirstDateSurpassedScript, &updateFirstDateSurpassed{
		ApplicationIDs:     firstDateSurpassed.ApplicationIDs,
		FirstDateSurpassed: firstDateSurpassed.FirstDateSurpassed,
		UpdatedAt:          time.Now(),
	})
	if err != nil {
		return err
	}

	query, args, err = sqlx.In(query, args...)
	if err != nil {
		return err
	}

	query = d.Rebind(query)

	_, err = d.Exec(query, args...)
	if err != nil {
		return err
	}

	return nil
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
