package entity

import "time"

type BaseTimestamps struct {
	CreatedAt time.Time  `bun:"created_at,default:current_timestamp"`
	UpdatedAt *time.Time `bun:"updated_at,default:current_timestamp"`
}
