package entity

import "github.com/uptrace/bun"

type Occupant struct {
	bun.BaseModel `bun:"table:occupant"`
	Id            int64        `bun:"id,pk,autoincrement"`
	Role          OccupantRole `bun:"role"`
	HouseId       int64        `bun:"house_id"`
	Name          string       `bun:"name"`
	Email         *string      `bun:"email"`
	Phone         string       `bun:"phone"`
	Password      string       `bun:"password"`
	BaseTimestamps
}

type OccupantRole string

const (
	OccupantRole_Owner  OccupantRole = "owner"
	OccupantRole_Renter OccupantRole = "renter"
)
