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

package rule

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/ory/ladon/compiler"
	"github.com/ory/oathkeeper/helper"
	"github.com/pkg/errors"
	"github.com/rubenv/sql-migrate"
)

type sqlRule struct {
	ID               string `db:"surrogate_id"`
	InternalID       int64  `db:"id"`
	MatchesMethods   string `db:"matches_methods"`
	MatchesURL       string `db:"matches_url"`
	RequiredScopes   string `db:"required_scopes"`
	RequiredAction   string `db:"required_action"`
	RequiredResource string `db:"required_resource"`
	Description      string `db:"description"`
	Mode             string `db:"mode"`
}

func (r *sqlRule) toRule() (*Rule, error) {
	exp, err := compiler.CompileRegex(r.MatchesURL, '<', '>')
	if err != nil {
		return nil, errors.WithStack(err)
	}

	methods := []string{}
	scopes := []string{}
	if len(r.MatchesMethods) > 0 {
		methods = strings.Split(r.MatchesMethods, " ")
	}
	if len(r.RequiredScopes) > 0 {
		scopes = strings.Split(r.RequiredScopes, " ")
	}

	return &Rule{
		ID:                 r.ID,
		MatchesMethods:     methods,
		MatchesURLCompiled: exp,
		MatchesURL:         r.MatchesURL,
		RequiredScopes:     scopes,
		RequiredAction:     r.RequiredAction,
		RequiredResource:   r.RequiredResource,
		Mode:               r.Mode,
		Description:        r.Description,
	}, nil
}

func toSqlRule(r *Rule) *sqlRule {
	return &sqlRule{
		ID:               r.ID,
		MatchesMethods:   strings.Join(r.MatchesMethods, " "),
		MatchesURL:       r.MatchesURL,
		RequiredScopes:   strings.Join(r.RequiredScopes, " "),
		RequiredAction:   r.RequiredAction,
		RequiredResource: r.RequiredResource,
		Description:      r.Description,
		Mode:             r.Mode,
	}
}

var migrations = &migrate.MemoryMigrationSource{
	Migrations: []*migrate.Migration{
		{
			Id: "1",
			Up: []string{`CREATE TABLE IF NOT EXISTS oathkeeper_rule (
	id      			SERIAL PRIMARY KEY,
	surrogate_id      	varchar(190) NOT NULL UNIQUE,
	matches_methods		varchar(64) NOT NULL,
	matches_url			text NOT NULL,
	required_scopes		text NOT NULL,
	required_action		text NOT NULL,
	required_resource	text NOT NULL,
	description			text NOT NULL,
	mode				text NOT NULL
)`},
			Down: []string{
				"DROP TABLE hydra_client",
			},
		},
	},
}

var sqlParams = []string{
	"surrogate_id",
	"matches_methods",
	"matches_url",
	"required_scopes",
	"required_action",
	"required_resource",
	"description",
	"mode",
}

func NewSQLManager(db *sqlx.DB) *SQLManager {
	return &SQLManager{db: db}
}

type SQLManager struct {
	db *sqlx.DB
}

func (s *SQLManager) CreateSchemas() (int, error) {
	migrate.SetTable("oathkeeper_rule_migration")
	n, err := migrate.Exec(s.db.DB, s.db.DriverName(), migrations, migrate.Up)
	if err != nil {
		return 0, errors.Wrapf(err, "Could not migrate sql schema, applied %d migrations", n)
	}
	return n, nil
}

func (s *SQLManager) ListRules() ([]Rule, error) {
	var d []sqlRule
	if err := s.db.Select(&d, "SELECT * FROM oathkeeper_rule"); err == sql.ErrNoRows {
		return []Rule{}, nil
	} else if err != nil {
		return nil, errors.WithStack(err)
	}

	rules := make([]Rule, len(d))
	for k, r := range d {
		rule, err := r.toRule()
		if err != nil {
			return nil, errors.WithStack(err)
		}
		rules[k] = *rule
	}

	return rules, nil
}

func (s *SQLManager) GetRule(id string) (*Rule, error) {
	var d sqlRule
	if err := s.db.Get(&d, s.db.Rebind("SELECT * FROM oathkeeper_rule WHERE surrogate_id=?"), id); err == sql.ErrNoRows {
		return nil, errors.WithStack(helper.ErrResourceNotFound)
	} else if err != nil {
		return nil, errors.WithStack(err)
	}

	return d.toRule()
}

func (s *SQLManager) CreateRule(rule *Rule) error {
	sr := toSqlRule(rule)

	if _, err := s.db.NamedExec(fmt.Sprintf(
		"INSERT INTO oathkeeper_rule (%s) VALUES (%s)",
		strings.Join(sqlParams, ", "),
		":"+strings.Join(sqlParams, ", :"),
	), sr); err != nil {
		return errors.WithStack(err)
	}
	return nil

}

func (s *SQLManager) UpdateRule(rule *Rule) error {
	sr := toSqlRule(rule)
	var query []string
	for _, param := range sqlParams {
		query = append(query, fmt.Sprintf("%s=:%s", param, param))
	}
	if _, err := s.db.NamedExec(fmt.Sprintf(`UPDATE oathkeeper_rule SET %s WHERE surrogate_id=:surrogate_id`, strings.Join(query, ", ")), sr); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *SQLManager) DeleteRule(id string) error {
	if _, err := s.db.Exec(s.db.Rebind(`DELETE FROM oathkeeper_rule WHERE surrogate_id=?`), id); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
