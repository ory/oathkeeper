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

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	migrate "github.com/rubenv/sql-migrate"

	"github.com/ory/oathkeeper/helper"
)

func NewSQLManager(db *sqlx.DB) *SQLManager {
	return &SQLManager{db: db}
}

type SQLManager struct {
	db *sqlx.DB
}

func (s *SQLManager) CreateSchemas() (int, error) {
	migrate.SetTable("oathkeeper_rule_migration")
	n, err := migrate.Exec(s.db.DB, s.db.DriverName(), migrations(s.db.DriverName()), migrate.Up)
	if err != nil {
		return 0, errors.Wrapf(err, "Could not migrate sql schema, applied %d migrations", n)
	}
	return n, nil
}

func (s *SQLManager) ListRules(limit, offset int) ([]Rule, error) {
	var ids []string
	if err := s.db.Select(&ids, s.db.Rebind("SELECT surrogate_id FROM oathkeeper_rule ORDER BY id LIMIT ? OFFSET ?"), limit, offset); err == sql.ErrNoRows {
		return []Rule{}, nil
	} else if err != nil {
		return nil, errors.WithStack(err)
	}

	rules := make([]Rule, len(ids))
	for k, id := range ids {
		d, err := s.GetRule(id)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		rules[k] = *d
	}

	return rules, nil
}

func (s *SQLManager) GetRule(id string) (*Rule, error) {
	d := sqlRule{
		Authenticators:     []sqlRuleHandler{},
		Authorizers:        []sqlRuleHandler{},
		CredentialsIssuers: []sqlRuleHandler{},
	}
	if err := s.db.Get(&d, s.db.Rebind("SELECT * FROM oathkeeper_rule WHERE surrogate_id=?"), id); err == sql.ErrNoRows {
		return nil, errors.WithStack(helper.ErrResourceNotFound)
	} else if err != nil {
		return nil, errors.WithStack(err)
	}

	if err := s.getHandler("oathkeeper_rule_authorizer", id, &d.Authorizers); err != nil {
		return nil, errors.WithStack(err)
	}
	if err := s.getHandler("oathkeeper_rule_authenticator", id, &d.Authenticators); err != nil {
		return nil, errors.WithStack(err)
	}
	if err := s.getHandler("oathkeeper_rule_credentials_issuer", id, &d.CredentialsIssuers); err != nil {
		return nil, errors.WithStack(err)
	}

	return d.toRule()
}

func (s *SQLManager) getHandler(table, id string, rh *sqlRuleHandlers) (err error) {
	if err := s.db.Select(rh, s.db.Rebind(fmt.Sprintf("SELECT * FROM %s WHERE rule_id=?", table)), id); err != nil {
		return errors.WithStack(err)
	}
	return
}

func (s *SQLManager) CreateRule(rule *Rule) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return errors.WithStack(err)
	}

	if err := s.createRule(tx, rule); err != nil {
		return err
	}

	return commit(tx)
}

func (s *SQLManager) createRule(tx *sqlx.Tx, rule *Rule) error {
	sr := toSqlRule(rule)
	if err := s.create(tx, "oathkeeper_rule", sqlParams, sr, "surrogate_id"); err != nil {
		return err
	}
	if err := s.createMany(tx, "oathkeeper_rule_authorizer", sqlParamsHandler, sr.Authorizers, rule.ID); err != nil {
		return err
	}
	if err := s.createMany(tx, "oathkeeper_rule_authenticator", sqlParamsHandler, sr.Authenticators, rule.ID); err != nil {
		return err
	}
	if err := s.createMany(tx, "oathkeeper_rule_credentials_issuer", sqlParamsHandler, sr.CredentialsIssuers, rule.ID); err != nil {
		return err
	}
	return nil
}

func (s *SQLManager) createMany(tx *sqlx.Tx, table string, params []string, value []sqlRuleHandler, id string) error {
	if _, err := tx.Exec(s.db.Rebind(fmt.Sprintf(`DELETE FROM %s WHERE rule_id=?`, table)), id); err != nil {
		return errors.WithStack(err)
	}

	for _, v := range value {
		if err := s.create(tx, table, params, v, "id"); err != nil {
			return err
		}
	}
	return nil
}

func (s *SQLManager) create(tx *sqlx.Tx, table string, params []string, value interface{}, primaryKey string) error {
	updateSet := ""
	for _, p := range params {
		updateSet += p + " = excluded." + p + ", "
	}
	updateSet = updateSet[:len(updateSet)-2]

	createQuery := "INSERT INTO " + table + " (%s) VALUES (%s) ON CONFLICT (" + primaryKey + ") DO UPDATE SET " + updateSet
	if s.db.DriverName() == "mysql" {
		createQuery = "REPLACE INTO " + table + " (%s) VALUES (%s)"
	}

	if _, err := tx.NamedExec(s.db.Rebind(fmt.Sprintf(
		createQuery,
		strings.Join(params, ", "),
		":"+strings.Join(params, ", :"),
	)), value); err != nil {
		if rErr := tx.Rollback(); rErr != nil {
			return errors.Wrap(rErr, err.Error())
		}
		return errors.WithStack(err)
	}
	return nil
}

func (s *SQLManager) UpdateRule(rule *Rule) error {
	_, err := s.GetRule(rule.ID)
	if errors.Cause(err) == helper.ErrResourceNotFound {
		return errors.WithStack(helper.ErrResourceConflict)
	} else if err != nil {
		return err
	}

	tx, err := s.db.Beginx()
	if err != nil {
		return errors.WithStack(err)
	}

	if err := s.deleteRule(tx, rule.ID); err != nil {
		return err
	}

	if err := s.createRule(tx, rule); err != nil {
		return err
	}

	return commit(tx)
}

func commit(tx *sqlx.Tx) error {
	if err := tx.Commit(); err != nil {
		if rErr := tx.Rollback(); rErr != nil {
			return errors.Wrap(rErr, err.Error())
		}
		return errors.WithStack(err)
	}
	return nil
}

func (s *SQLManager) DeleteRule(id string) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return errors.WithStack(err)
	}

	if err := s.deleteRule(tx, id); err != nil {
		return err
	}

	return commit(tx)
}

func (s *SQLManager) deleteRule(tx *sqlx.Tx, id string) error {
	if _, err := tx.Exec(s.db.Rebind(`DELETE FROM oathkeeper_rule WHERE surrogate_id=?`), id); err != nil {
		if rErr := tx.Rollback(); rErr != nil {
			return errors.Wrap(rErr, err.Error())
		}
		return errors.WithStack(err)
	}
	return nil
}
