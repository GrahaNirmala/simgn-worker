package entity

import (

	"github.com/uptrace/bun"
)

type BillingConfig struct {
	bun.BaseModel `bun:"table:billing_config"`
	Id            	  int64     `bun:"id,pk,autoincrement"`
	AmountBill        int64     `bun:"amount_bill"`
	ExtraChargeBill   int64     `bun:"extra_charge_bill"`
	BaseTimestamps
}
