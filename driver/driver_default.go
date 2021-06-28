// Copyright 2021 Ory GmbH
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package driver

import (
	"github.com/ory/x/logrusx"

	"github.com/ory/oathkeeper/driver/configuration"
)

type DefaultDriver struct {
	c configuration.Provider
	r Registry
}

func NewDefaultDriver(l *logrusx.Logger, version, build, date string) Driver {
	c := configuration.NewViperProvider(l)
	r := NewRegistry(c).WithLogger(l).WithBuildInfo(version, build, date)

	return &DefaultDriver{r: r, c: c}
}

func (r *DefaultDriver) Configuration() configuration.Provider {
	return r.c
}

func (r *DefaultDriver) Registry() Registry {
	return r.r
}
