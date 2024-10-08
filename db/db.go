package db

import (
	"database/sql"
	"fmt"

	"github.com/Roofiif/sim-graha-nirmala-worker/config"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

type Client struct {
	db *bun.DB
}

func NewClient() (*Client, error) {
	dataSourceName := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=require",
		config.Cfg().Db.User,
		config.Cfg().Db.Password,
		config.Cfg().Db.Host,
		config.Cfg().Db.Database,
	)

	sqlDb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dataSourceName)))
	bunDb := bun.NewDB(sqlDb, pgdialect.New())
	err := bunDb.Ping()
	if err != nil {
		return nil, err
	}

	return &Client{
		db: bunDb,
	}, nil
}

func (c *Client) Conn() *bun.DB { return c.db }
func (c *Client) Close() error  { return c.db.Close() }
