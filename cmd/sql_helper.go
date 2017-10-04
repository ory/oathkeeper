package cmd

import (
	"runtime"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"net/url"
	"github.com/ory/oathkeeper/rule"
)

func connectToSql(url string) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", url)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	maxConns := maxParallelism() * 2
	maxConnLifetime := time.Duration(0)
	maxIdleConns := maxParallelism()
	db.SetMaxOpenConns(maxConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(maxConnLifetime)
	return db, nil
}

func maxParallelism() int {
	maxProcs := runtime.GOMAXPROCS(0)
	numCPU := runtime.NumCPU()
	if maxProcs < numCPU {
		return maxProcs
	}
	return numCPU
}

func newRuleManager(db string) (rule.Manager, error) {
	if db == "memory" {
		return &rule.MemoryManager{Rules: map[string]Rule{}}, nil
	} else if db == "" {
		return nil, errors.New("No database URL provided")
	}

	u, err := url.Parse(db)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	switch u.Scheme {
	case "postgres":
		db, err := connectToSql(db)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		return rule.NewSQLManager(db), nil
	}

	return nil, errors.Errorf("The provided database URL %s can not be handled", db)
}
