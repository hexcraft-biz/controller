package controller

import (
	"github.com/jmoiron/sqlx"
)

type Prototype struct {
	Name string
	DB   *sqlx.DB
}

func New(name string, db *sqlx.DB) *Prototype {
	return &Prototype{
		Name: name,
		DB:   db,
	}
}
