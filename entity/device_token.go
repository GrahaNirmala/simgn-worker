package entity

import "github.com/uptrace/bun"

type DeviceToken struct {
	bun.BaseModel `bun:"table:device_token"`
	Id            int64  `bun:"id,pk,autoincrement"`
	OccupantId    int64  `bun:"occupant_id"`
	DeviceToken   string `bun:"device_token"`
	BaseTimestamps
}
