/*
 * Copyright © 2017-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * @author       Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @copyright  2017-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @license  	   Apache-2.0
 */

package cmd

import (
	"runtime"
	"strings"
	"time"

	"net/url"

	"github.com/jmoiron/sqlx"
	"github.com/ory/oathkeeper/rule"
	"github.com/pkg/errors"
)

func cleanURLQuery(c *url.URL) *url.URL {
	cleanurl := new(url.URL)
	*cleanurl = *c

	q := cleanurl.Query()
	q.Del("max_conns")
	q.Del("max_idle_conns")
	q.Del("max_conn_lifetime")

	cleanurl.RawQuery = q.Encode()
	return cleanurl
}

func connectToSql(url *url.URL) (*sqlx.DB, error) {
	clean := cleanURLQuery(url)
	u := clean.String()

	if clean.Scheme == "mysql" {
		u = strings.Replace(u, "mysql://", "", -1)
	}

	db, err := sqlx.Open(clean.Scheme, u)
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
		return &rule.MemoryManager{Rules: map[string]rule.Rule{}}, nil
	} else if db == "" {
		return nil, errors.New("No database URL provided")
	}

	u, err := url.Parse(db)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	switch u.Scheme {
	case "postgres":
	case "mysql":
		db, err := connectToSql(u)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		return rule.NewSQLManager(db), nil
	}

	return nil, errors.Errorf("The provided database URL %s can not be handled", db)
}
