package postgresdriver

import "github.com/pokt-foundation/portal-api-go/repository"

const (
	selectPayPlans = "SELECT plan_type, daily_limit FROM pay_plans"
)

type dbPayPlan struct {
	PlanType   string `db:"plan_type"`
	DailyLimit int    `db:"daily_limit"`
}

func (d dbPayPlan) toPayPlan() repository.PayPlan {
	return repository.PayPlan{
		PlanType:   repository.PayPlanType(d.PlanType),
		DailyLimit: d.DailyLimit,
	}
}

// ReadPayPlans returns all pay plans on the database
func (d *PostgresDriver) ReadPayPlans() ([]repository.PayPlan, error) {
	var dbPayPlans []dbPayPlan

	err := d.Select(&dbPayPlans, selectPayPlans)
	if err != nil {
		return nil, err
	}

	var payPlans []repository.PayPlan

	for _, dbPayPlan := range dbPayPlans {
		payPlans = append(payPlans, dbPayPlan.toPayPlan())
	}

	return payPlans, nil
}
