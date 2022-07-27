package postgresdriver

import (
	"database/sql"
	"strings"

	"github.com/lib/pq"
	"github.com/pokt-foundation/portal-api-go/repository"
)

const (
	selectLoadBalancers = `
	SELECT lb.lb_id, lb.name, lb.created_at, lb.updated_at, lb.request_timeout, lb.gigastake, lb.gigastake_redirect, lb.user_id, so.duration, so.relays_limit, so.stickiness, so.temp, so.origins, so.use_rpc_id, ARRAY_AGG(la.app_id) AS app_ids
	FROM loadbalancers AS lb
	LEFT JOIN stickiness_options AS so ON lb.lb_id=so.lb_id
	LEFT JOIN lb_apps AS la ON lb.lb_id=la.lb_id
	GROUP BY lb.lb_id, lb.lb_id, lb.name, lb.created_at, lb.updated_at, lb.request_timeout, lb.gigastake, lb.gigastake_redirect, lb.user_id, so.duration, so.relays_limit, so.stickiness, so.temp, so.origins, so.use_rpc_id
	`
)

type dbLoadBalancer struct {
	LbID              string         `db:"lb_id"`
	Name              sql.NullString `db:"name"`
	CreatedAt         sql.NullTime   `db:"created_at"`
	UpdatedAt         sql.NullTime   `db:"updated_at"`
	RequestTimeout    sql.NullInt64  `db:"request_timeout"`
	Gigastake         sql.NullBool   `db:"gigastake"`
	GigastakeRedirect sql.NullBool   `db:"gigastake_redirect"`
	UserID            sql.NullString `db:"user_id"`
	Duration          sql.NullString `db:"duration"`
	RelaysLimit       sql.NullInt64  `db:"relays_limit"`
	Stickiness        sql.NullBool   `db:"stickiness"`
	Temp              sql.NullString `db:"temp"`
	Origins           pq.StringArray `db:"origins"`
	UseRPCID          sql.NullBool   `db:"use_rpc_id"`
	AppIDS            sql.NullString `db:"app_ids"`
}

func (lb *dbLoadBalancer) toLoadBalancer() *repository.LoadBalancer {
	return &repository.LoadBalancer{
		ID:                lb.LbID,
		Name:              lb.Name.String,
		CreatedAt:         lb.CreatedAt.Time,
		UpdatedAt:         lb.UpdatedAt.Time,
		RequestTimeout:    lb.RequestTimeout.Int64,
		Gigastake:         lb.Gigastake.Bool,
		GigastakeRedirect: lb.GigastakeRedirect.Bool,
		UserID:            lb.UserID.String,
		ApplicationIDs:    strings.Split(lb.AppIDS.String, ","),
		StickyOptions: repository.StickyOptions{
			Duration:      lb.Duration.String,
			RelaysLimit:   int(lb.RelaysLimit.Int64),
			Stickiness:    lb.Stickiness.Bool,
			Temp:          lb.Temp.String,
			StickyOrigins: lb.Origins,
			UseRPCID:      lb.UseRPCID.Bool,
		},
	}
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
