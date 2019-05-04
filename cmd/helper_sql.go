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
	"net/url"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/ory/oathkeeper/rule"
	"github.com/ory/x/sqlcon"
)

func connectToSql(dburl string) (*sqlx.DB, error) {
	u, err := url.Parse(dburl)
	if err != nil {
		logger.Fatalf("Could not parse DATABASE_URL: %s", err)
	}

	switch u.Scheme {
	case "postgres":
		fallthrough
	case "mysql":
		connection, err := sqlcon.NewSQLConnection(dburl, logger)
		if err != nil {
			logger.WithError(err).Fatalf(`Unable to initialize SQL connection`)
		}
		return connection.GetDatabase()
	}

	return nil, errors.Errorf(`Unknown DSN "%s" in DATABASE_URL: %s`, u.Scheme, dburl)
}

func connectToDatabase(dburl string) (interface{}, error) {
	if dburl == "memory" {
		return nil, nil
	} else if dburl == "" {
		return nil, errors.New("No database URL provided")
	}

	db, err := connectToSql(dburl)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return db, nil
}

func newRuleManager(database interface{}) (rule.Manager, error) {
	if database == nil {
		return &rule.MemoryManager{Rules: map[string]rule.Rule{}}, nil
	}

	switch db := database.(type) {
	case *sqlx.DB:
		return rule.NewSQLManager(db), nil
	default:
		return nil, errors.New("Unknown database type")
	}
}
