package postgresdriver

import (
	"database/sql"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/pokt-foundation/portal-api-go/repository"
)

const (
	selectLoadBalancers = `
	SELECT lb.lb_id, lb.name, lb.created_at, lb.updated_at, lb.request_timeout, lb.gigastake, lb.gigastake_redirect, lb.user_id, so.duration, so.relays_limit, so.stickiness, so.origins, so.use_rpc_id, STRING_AGG(la.app_id, ',') AS app_ids
	FROM loadbalancers AS lb
	LEFT JOIN stickiness_options AS so ON lb.lb_id=so.lb_id
	LEFT JOIN lb_apps AS la ON lb.lb_id=la.lb_id
	GROUP BY lb.lb_id, lb.lb_id, lb.name, lb.created_at, lb.updated_at, lb.request_timeout, lb.gigastake, lb.gigastake_redirect, lb.user_id, so.duration, so.relays_limit, so.stickiness, so.temp, so.origins, so.use_rpc_id
	`
	insertLoadBalancerScript = `
	INSERT into loadbalancers (lb_id, name, user_id, request_timeout, gigastake, gigastake_redirect, created_at, updated_at)
	VALUES (:lb_id, :name, :user_id, :request_timeout, :gigastake, :gigastake_redirect, :created_at, :updated_at)`
	insertLbAppsScript = `
	INSERT into lb_apps (lb_id, app_id)
	VALUES (:lb_id, :app_id)`
	updateLoadBalancer = `
	UPDATE loadbalancers
	SET name = COALESCE($1, name), user_id = COALESCE($2, user_id), updated_at = $3
	WHERE lb_id = $4`
)

type dbLoadBalancer struct {
	LbID              string         `db:"lb_id"`
	Duration          sql.NullString `db:"duration"`
	Name              sql.NullString `db:"name"`
	UserID            sql.NullString `db:"user_id"`
	AppIDs            sql.NullString `db:"app_ids"`
	Origins           pq.StringArray `db:"origins"`
	RelaysLimit       sql.NullInt32  `db:"relays_limit"`
	RequestTimeout    sql.NullInt32  `db:"request_timeout"`
	Gigastake         sql.NullBool   `db:"gigastake"`
	GigastakeRedirect sql.NullBool   `db:"gigastake_redirect"`
	Stickiness        sql.NullBool   `db:"stickiness"`
	UseRPCID          sql.NullBool   `db:"use_rpc_id"`
	CreatedAt         sql.NullTime   `db:"created_at"`
	UpdatedAt         sql.NullTime   `db:"updated_at"`
}

func getAppIDs(rawAppIDs sql.NullString) []string {
	if !rawAppIDs.Valid {
		return nil
	}

	appIDs := strings.Split(rawAppIDs.String, ",")

	// This is needed in some cases where appIDs were not stored correctly
	if strings.Contains(appIDs[0], "{$oid:") {
		for i := 0; i < len(appIDs); i++ {
			appIDs[i] = appIDs[i][6 : len(appIDs[i])-1]
		}
	}

	return appIDs
}

func (lb *dbLoadBalancer) toLoadBalancer() *repository.LoadBalancer {
	return &repository.LoadBalancer{
		ID:                lb.LbID,
		Name:              lb.Name.String,
		UserID:            lb.UserID.String,
		ApplicationIDs:    getAppIDs(lb.AppIDs),
		RequestTimeout:    int(lb.RequestTimeout.Int32),
		Gigastake:         lb.Gigastake.Bool,
		GigastakeRedirect: lb.GigastakeRedirect.Bool,
		StickyOptions: repository.StickyOptions{
			Duration:      lb.Duration.String,
			StickyOrigins: lb.Origins,
			RelaysLimit:   int(lb.RelaysLimit.Int32),
			Stickiness:    lb.Stickiness.Bool,
			UseRPCID:      lb.UseRPCID.Bool,
		},
		CreatedAt: lb.CreatedAt.Time,
		UpdatedAt: lb.UpdatedAt.Time,
	}
}

type insertLoadBalancer struct {
	LbID              string         `db:"lb_id"`
	Name              sql.NullString `db:"name"`
	UserID            sql.NullString `db:"user_id"`
	RequestTimeout    sql.NullInt32  `db:"request_timeout"`
	Gigastake         bool           `db:"gigastake"`
	GigastakeRedirect bool           `db:"gigastake_redirect"`
	CreatedAt         time.Time      `db:"created_at"`
	UpdatedAt         time.Time      `db:"updated_at"`
}

func extractInsertLoadBalancer(loadBalancer *repository.LoadBalancer) *insertLoadBalancer {
	return &insertLoadBalancer{
		LbID:              loadBalancer.ID,
		Name:              newSQLNullString(loadBalancer.Name),
		UserID:            newSQLNullString(loadBalancer.UserID),
		RequestTimeout:    newSQLNullInt32(int32(loadBalancer.RequestTimeout)),
		Gigastake:         loadBalancer.Gigastake,
		GigastakeRedirect: loadBalancer.GigastakeRedirect,
		CreatedAt:         loadBalancer.CreatedAt,
		UpdatedAt:         loadBalancer.UpdatedAt,
	}
}

type insertLbApps struct {
	LbID  string `db:"lb_id"`
	AppID string `db:"app_id"`
}

func extractInsertLbApps(loadBalancer *repository.LoadBalancer) []*insertLbApps {
	var inserts []*insertLbApps

	for _, appID := range loadBalancer.ApplicationIDs {
		inserts = append(inserts, &insertLbApps{
			LbID:  loadBalancer.ID,
			AppID: appID,
		})
	}

	return inserts
}

// ReadLoadBalancers returns all load balancers in the database
func (d *PostgresDriver) ReadLoadBalancers() ([]*repository.LoadBalancer, error) {
	var dbLoadBalancers []*dbLoadBalancer

	err := d.Select(&dbLoadBalancers, selectLoadBalancers)
	if err != nil {
		return nil, err
	}

	var loadbalancers []*repository.LoadBalancer

	for _, dbLoadBalancer := range dbLoadBalancers {
		loadbalancers = append(loadbalancers, dbLoadBalancer.toLoadBalancer())
	}

	return loadbalancers, nil
}

// WriteLoadBalancer saves input load balancer in the database
// Does not save stickiness configuration
func (d *PostgresDriver) WriteLoadBalancer(loadBalancer *repository.LoadBalancer) (*repository.LoadBalancer, error) {
	id, err := generateRandomID()
	if err != nil {
		return nil, err
	}

	loadBalancer.ID = id
	loadBalancer.CreatedAt = time.Now()
	loadBalancer.UpdatedAt = time.Now()

	insertLoadBalancer := extractInsertLoadBalancer(loadBalancer)
	insertsLbApps := extractInsertLbApps(loadBalancer)

	tx, err := d.Beginx()
	if err != nil {
		return nil, err
	}

	_, err = tx.NamedExec(insertLoadBalancerScript, insertLoadBalancer)
	if err != nil {
		return nil, err
	}

	for _, insert := range insertsLbApps {
		_, err = tx.NamedExec(insertLbAppsScript, insert)
		if err != nil {
			return nil, err
		}
	}

	return loadBalancer, tx.Commit()
}

// UpdateLoadBalancer updates fields available in options in db
func (d *PostgresDriver) UpdateLoadBalancer(id string, fieldsToUpdate *repository.UpdateLoadBalancer) error {
	if id == "" {
		return ErrMissingID
	}

	if fieldsToUpdate == nil {
		return ErrNoFieldsToUpdate
	}

	_, err := d.Exec(updateLoadBalancer, newSQLNullString(fieldsToUpdate.Name),
		newSQLNullString(fieldsToUpdate.UserID), time.Now(), id)
	if err != nil {
		return err
	}

	return nil
}
