package rule

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/ory/ladon/compiler"
	"github.com/ory/oathkeeper/helper"
	"github.com/pkg/errors"
	"github.com/rubenv/sql-migrate"
)

type sqlRule struct {
	ID                          string `db:"id"`
	MatchesMethods              string `db:"matches_methods"`
	MatchesURL                  string `db:"matches_url"`
	RequiredScopes              string `db:"required_scopes"`
	RequiredAction              string `db:"required_action"`
	RequiredResource            string `db:"required_resource"`
	AllowAnonymous              bool   `db:"allow_anonymous"`
	BypassAuthorization         bool   `db:"disable_firewall"`
	BypassAccessControlPolicies bool   `db:"disable_acp"`
	Description                 string `db:"description"`
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
		ID:                          r.ID,
		MatchesMethods:              methods,
		MatchesURLCompiled:          exp,
		MatchesURL:                  r.MatchesURL,
		RequiredScopes:              scopes,
		RequiredAction:              r.RequiredAction,
		RequiredResource:            r.RequiredResource,
		AllowAnonymous:              r.AllowAnonymous,
		BypassAuthorization:         r.BypassAuthorization,
		BypassAccessControlPolicies: r.BypassAccessControlPolicies,
		Description:                 r.Description,
	}, nil
}

func toSqlRule(r *Rule) *sqlRule {
	return &sqlRule{
		ID:                          r.ID,
		MatchesMethods:              strings.Join(r.MatchesMethods, " "),
		MatchesURL:                  r.MatchesURL,
		RequiredScopes:              strings.Join(r.RequiredScopes, " "),
		RequiredAction:              r.RequiredAction,
		RequiredResource:            r.RequiredResource,
		AllowAnonymous:              r.AllowAnonymous,
		BypassAuthorization:         r.BypassAuthorization,
		BypassAccessControlPolicies: r.BypassAccessControlPolicies,
		Description:                 r.Description,
	}
}

var migrations = &migrate.MemoryMigrationSource{
	Migrations: []*migrate.Migration{
		{
			Id: "1",
			Up: []string{`CREATE TABLE IF NOT EXISTS oathkeeper_rule (
	id      			varchar(64) NOT NULL PRIMARY KEY,
	matches_methods		varchar(64) NOT NULL,
	matches_url		text NOT NULL,
	required_scopes		text NOT NULL,
	required_action		text NOT NULL,
	required_resource	text NOT NULL,
	allow_anonymous		bool NOT NULL,
	description			text NOT NULL,
	disable_firewall	bool NOT NULL,
	disable_acp			bool NOT NULL
)`},
			Down: []string{
				"DROP TABLE hydra_client",
			},
		},
	},
}

var sqlParams = []string{
	"id",
	"matches_methods",
	"matches_url",
	"required_scopes",
	"required_action",
	"required_resource",
	"allow_anonymous",
	"description",
	"disable_firewall",
	"disable_acp",
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
	if err := s.db.Get(&d, s.db.Rebind("SELECT * FROM oathkeeper_rule WHERE id=?"), id); err == sql.ErrNoRows {
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
	if _, err := s.db.NamedExec(fmt.Sprintf(`UPDATE oathkeeper_rule SET %s WHERE id=:id`, strings.Join(query, ", ")), sr); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *SQLManager) DeleteRule(id string) error {
	if _, err := s.db.Exec(s.db.Rebind(`DELETE FROM oathkeeper_rule WHERE id=?`), id); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
