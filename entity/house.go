package entity

import "github.com/uptrace/bun"

type House struct {
	bun.BaseModel `bun:"table:house"`
	Id            int64  `bun:"id,pk,autoincrement"`
	Code          string `bun:"code"`
	Address       string `bun:"address"`
	BaseTimestamps
}
