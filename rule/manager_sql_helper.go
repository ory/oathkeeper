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
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/rubenv/sql-migrate"
)

type sqlRule struct {
	ID                   string `db:"surrogate_id"`
	InternalID           int64  `db:"id"`
	MatchesMethods       string `db:"matches_methods"`
	MatchesURL           string `db:"matches_url"`
	Description          string `db:"description"`
	UpstreamURL          string `db:"upstream_url"`
	UpstreamPreserveHost bool   `db:"upstream_preserve_host"`
	UpstreamStripPath    string `db:"upstream_strip_path"`

	Authorizers        sqlRuleHandlers
	CredentialsIssuers sqlRuleHandlers
	Authenticators     sqlRuleHandlers
}
type sqlRuleHandlers []sqlRuleHandler
type sqlRuleHandler struct {
	ID      string `db:"id"`
	Rule    string `db:"rule_id"`
	Handler string `db:"handler"`
	Config  string `db:"config"`
}

func (s sqlRuleHandlers) toRuleHandler() []RuleHandler {
	rh := make([]RuleHandler, len(s))
	for k, v := range s {
		rh[k] = RuleHandler{
			Handler: v.Handler,
			Config:  []byte(v.Config),
		}
	}
	return rh
}

func (r *sqlRule) toRule() (*Rule, error) {

	fmt.Printf("%+v", r)

	methods := []string{}
	if len(r.MatchesMethods) > 0 {
		methods = strings.Split(r.MatchesMethods, " ")
	}

	if len(r.Authorizers) > 1 {
		return nil, errors.New("Expected at most one oathkeeper_rule_authorizer row, but found none")
	} else if len(r.Authorizers) == 0 {
		r.Authorizers = []sqlRuleHandler{{}}
	}

	if len(r.CredentialsIssuers) > 1 {
		return nil, errors.New("Expected at most one oathkeeper_rule_credentials_issuer row, but found none")
	} else if len(r.CredentialsIssuers) == 0 {
		r.CredentialsIssuers = []sqlRuleHandler{{}}
	}

	return &Rule{
		ID: r.ID,
		Match: RuleMatch{
			URL:     r.MatchesURL,
			Methods: methods,
		},
		Authenticators:    r.Authenticators.toRuleHandler(),
		Authorizer:        r.Authorizers.toRuleHandler()[0],
		CredentialsIssuer: r.CredentialsIssuers.toRuleHandler()[0],
		Description:       r.Description,
		Upstream: Upstream{
			URL:          r.UpstreamURL,
			PreserveHost: r.UpstreamPreserveHost,
			StripPath:    r.UpstreamStripPath,
		},
	}, nil
}

func toSqlRuleHandler(rs []RuleHandler, rule string) sqlRuleHandlers {
	srh := make([]sqlRuleHandler, len(rs))
	for k, v := range rs {
		srh[k] = sqlRuleHandler{
			Rule:    rule,
			Handler: v.Handler,
			Config:  string(v.Config),
		}
	}
	return srh
}

func toSqlRule(r *Rule) *sqlRule {
	an := toSqlRuleHandler(r.Authenticators, r.ID)
	ci := toSqlRuleHandler([]RuleHandler{r.CredentialsIssuer}, r.ID)
	az := toSqlRuleHandler([]RuleHandler{r.Authorizer}, r.ID)

	return &sqlRule{
		ID:                   r.ID,
		MatchesMethods:       strings.Join(r.Match.Methods, " "),
		MatchesURL:           r.Match.URL,
		Description:          r.Description,
		UpstreamURL:          r.Upstream.URL,
		UpstreamPreserveHost: r.Upstream.PreserveHost,
		UpstreamStripPath:    r.Upstream.StripPath,

		Authorizers:        az,
		CredentialsIssuers: ci,
		Authenticators:     an,
	}
}

func primaryKey(database string) string {
	if database == "mysql" {
		return "INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY"
	}
	return "SERIAL PRIMARY KEY"
}

func relation(table, database string) string {
	return fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
	id      	%s,
	rule_id		varchar(190) NOT NULL,
	handler		varchar(64) NOT NULL,
	config		text NOT NULL,
	FOREIGN KEY (rule_id) REFERENCES oathkeeper_rule (surrogate_id) ON DELETE CASCADE
)`, table, primaryKey(database))
}

var migrations = func(database string) *migrate.MemoryMigrationSource {
	return &migrate.MemoryMigrationSource{
		Migrations: []*migrate.Migration{
			{
				Id: "1",
				Up: []string{
					fmt.Sprintf(`CREATE TABLE IF NOT EXISTS oathkeeper_rule (
	id      				%s,
	surrogate_id      		varchar(190) NOT NULL UNIQUE,
	matches_methods			varchar(128) NOT NULL,
	matches_url				text NOT NULL,
	upstream_url			text NOT NULL,
	upstream_preserve_host	bool NOT NULL,
	upstream_strip_path		text NOT NULL,
	description				text NOT NULL
)`, primaryKey(database)),
					relation("oathkeeper_rule_authorizer", database),
					relation("oathkeeper_rule_authenticator", database),
					relation("oathkeeper_rule_credentials_issuer", database),
				},
				Down: []string{
					"DROP TABLE oathkeeper_rule",
					"DROP TABLE oathkeeper_rule_authorizer",
					"DROP TABLE oathkeeper_rule_auththenticator",
					"DROP TABLE oathkeeper_rule_credentials_issuer",
				},
			},
		},
	}
}

var sqlParams = []string{
	"surrogate_id",
	"matches_methods",
	"matches_url",
	"description",
	"upstream_url",
	"upstream_preserve_host",
	"upstream_strip_path",
}

var sqlParamsHandler = []string{
	"handler",
	"config",
	"rule_id",
}
