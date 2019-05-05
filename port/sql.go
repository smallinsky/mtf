package port

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type MySQL struct {
}

func NewMYSQL(s string) *MySQL {
	db, err := sqlx.Connect("mysql", s)
	if err != nil {
		panic(err)
	}
	db = db
	return nil
}
