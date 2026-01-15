// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

func IfThenElseString(c bool, thenVal, elseVal string) string {
	if c {
		return thenVal
	}
	return elseVal
}
