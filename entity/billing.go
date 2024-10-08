package entity

import (
	"time"

	"github.com/uptrace/bun"
)

type Billing struct {
	bun.BaseModel `bun:"table:billing"`
	Id            int64     `bun:"id,pk,autoincrement"`
	HouseId       int64     `bun:"house_id"`
	Period        time.Time `bun:"period"`
	Amount        int64     `bun:"amount"`
	IsPaid        bool      `bun:"is_paid"`
	ExtraCharge   int64     `bun:"extra_charge"`
	BaseTimestamps
}
