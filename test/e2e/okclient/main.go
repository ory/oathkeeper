// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/go-jose/go-jose/v3"
	"github.com/golang-jwt/jwt/v5"

	"github.com/ory/oathkeeper/x"
	"github.com/ory/x/cmdx"
	"github.com/ory/x/urlx"
)

const key = `{
  "keys": [
    {
      "kid": "01cf6f34-07a8-40b7-b536-5fede7e2323e",
      "use": "sig",
      "kty": "RSA",
      "alg": "RS256",
      "n": "nuNHpDXf246QkUT3_WQn7hhinYC-dJN6dg2Oy4nn2cyUPZShQwHIZ6PEZioWS6XyYzrJ-WHasuQSaqVABFKMmRdK3yOeDRaOlijTgFXDEN2EZFJzKbbWNw63wGmh_VkZiuGxiZzqu9ViGGoduVIW-d7SizHcIPfyoc8mpigBYJ56n-_tqfGWN5DgLA44G4ALA_VpBbVueOwvOqEGfSffWqdFEqStpkJ63Dau8xI3hpIa0VUDCLlERL0mf9kMDg0sd_uhHKU2Kuvr1GIJqGftEV6m7XtyqrugpQUPM40vNOUsDzDCO6-0Q2y23El8XfzFi0k_O6R9YQoPNVkHE4RVkLauq9Z76l_04oSFICpN9UdeIYUZjHUn6HGw_jS_9ZWC35qdvOUKoX1hVphlbfLUK6KgYF2ABzjqraW7YVMbBy_x8A7C7ZIrkuheGyCy9JA5KjXleluMB23h9985LQXjLLBZMedGySbFUoybjJa8eFOUBDw9_nY-XJjae9lo-6DzGpP8wvNVgFjum60VhyPTfXvzXdxhIoovH39HkX1ATzWS9-kWg5u6S3xFC6wS06d2-usMtzCOYfUJuWTho5Tgqztilrcq3tZF9L6cgWbyd1ULIQIcxSXAIzkrxIuxMI-rraIFkslS2A_4dF1ff9KvncIThj8hXMI74meG1WLvRRs",
      "e": "AQAB",
      "d": "FXhw_eep0Gl7b6X6POuD2dDBwrbbBbXIHpvGaArCodPbVFb5U6yyAA9JQuT9DkvmbTJMiL6IZxQayP57oBnnlehn3a9adDpQGkp6CiWMroLAmS3rEf_7AoWinwWnNi0MHpBRBV-G-Mrk7peoXJqTMEDEArtCG9Jlnyf2-Qz-4qeHuoUZgJV8zxVstYdWXaD33BkVkAfsXbrMxwdrcQ4qSH3B_7rxHD5vi8J38bDMgArOROtP9aXTa9aDlVMnJd7x22TNaKmKoFIxFAoLdA5XIrm_pOfBnwCrvKDqZPT3JBaz1XmpRZeArHvMWgg0Hh9CMog7WzvnFW3ekD7vjducFDOXWY6dTuOx-lNemNfMPE4kpDfdPDLMDDg9x5Otr-aOcqn95pGb1I8s261c5TXoyiLqYFGJQBnJxmP7PdhjJfkx_nyrCqzVkTJSeJ5195cPuPC1AlP6RTZwzaRFplaHdEmWSOEdrUVQF-YDUFdl86XVcdYwWdG6-C40TDzu7OPrUqygFKezhDxrJOYCkpcTZNyUeKEbUfftUA1XDBeAniJ70iCcS-5U8sLSeaowNuJyBugi-zEohReKPwj8BxVozcKSc8Aos6NJuxigVodH973bfIsl7eGaO8Ah4neqiNtEKik5M2TxsD1eY6zk0ukqaU3_Wyisz8BOR0dsy4sBwgE",
      "p": "zWCe29eBsKU_EfLbqaoqvubo8kOGkUikDZWt9zY_joq_4to06qTQ6X9PGIGaDMa6TZmtFJIC7iGgQf1C6169_TwzlOWuYGZ1wIccJsj0MVhO8zXLRN4D6wnW75h_hH8S0hQDXn99_tdiI4gGArBBvNiRK110NaksYjrA1i98JXJk7WEWR-YVWZjGSnowixaKa6me6ymUpPazG81MqWJ-jjEROdF4kz9CSRnyxCa-2YjJbMVbQfCvovVN1g9p9pkuSRTt9yKin09X3nRNLfDy2R45Otb2zMAp9nXmmYClCTPvhjb8-YlsILDXhHmXwdBd02S5u9sWeW1uqV9SnhMi4Q",
      "q": "xg0med2Z6csvqkb83BH7Ql4tV9LmNQknIPwZZAFl4925uKwoe9Te1CGKyeANoaidKeHWUpjeDA-Ugf1uFWE39ajLi9ulpO2j-xn5oWJPi8MGQvoPfmFvBkE7Yjw8w1zZx8ObeRhHmciOaa1bt7Xd6JCtPMCgZIUtLFG6DU7zuMeDs3qRY0dMAtJxN9b8IMZRL3gT4hI7tK1weXkSOXuZhnzTIY44SINUXVY2XzIe5xxnxptwoBzwEGwbczaWiqAy2NvuDSSR5SzF-JIkvVPfbjAhobzfeV9T2CuAPiDN9Y5644PBkpagdPwxcyGTfAVV48tw70WRhpvRsu-xQdDjew",
      "dp": "LaJEahDOjlOJWhGeYWqgKN7h78K1Sd7cJPCXQRDaum92B4_4phKNIPTavFU_x4r4pnl5DdMLt6HdHAyMLARXLseLppAKWP0rOOZMiQmpyLUYhc30Jo82S6laCs7VtrkNk4RC42Jsuo1dDwwQWdSUJsPwLbSMN8gpRoJLisvuR9vNNTmXW85x_ACIPtA7rQwLIbUEzLfmgWkXjxvk5tYtzKZ2b2Rt6DnsHpYXbSI20dsk9ng7uMEkJY9gBipSnyqWMELoRTt66u8UVSd-ZrDdJQUkLkDJgeFI8axs2rrM8OiLVkUtnLy-O91PZe4mnDgSQQBmjKk3qy93oUa-8sTJwQ",
      "dq": "chRlE1FYNNk7LYQSZtlct4_z4mCzBd2t0kwke_xqKmUvc3aHIz0s1Vg0z6_lajbrwJI7l_xB-wSGcJIAdQQ47aI7DOaKdYQFLsB5kEJGu6Ss2KudfRi2kQ3GHQHBpm63a1_7EDeyFpz1m12mNR3nIu0jPJlgSkaEDMFOwBe4P7l83Uc-s8b_u6hoWtfVBzP007kBCvmhtqMo5G-e1cmiV1tTakK3nN05HUJps0_1iP5NsJMcvr2scmkzLpxE_rjkURHt_1gPSckOh-32h6_mRVwoIzfeKdWFQEFeb2sJ9-YIV0EvoTZuWrRh74VppMiV_s7S0KBsAPOESOPUYKroPQ",
      "qi": "V8zs0xY2S6Vg6EvXTG27In7N8uREAs9Wo25mNgB1JgpsEZLwM37859ZVCLVXZJ_Z5l4JznZ_FhvH5TTKevto_2W8mhR8t--WOwqgHr3Rn_81JIG3QLUStZvuM5TutXX4xncjr0ZsZ03qaOsmGdVkU_ot9Q5b6fHSrX7OewVH5xo_M5PNPs8yE-2KHLiYz6NdrpcATFGUEsgEedfE6PZFhpDzYXgWMmZg8-FY-CoXZhEsCHxTd749lfy6e52tR4QilOIMhdVtFMao_KyQFn8vy1yHas50BKET6bpz5cjJkd-NocC0xbE0jZQV2Xy2QJlCXa3fWCsMZXCIwaSg2hH2fw"
    }
  ]
}
`

var tokenValid = func() string {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"sub": "foobar",
		"exp": time.Now().UTC().Add(time.Hour).Unix(),
	})
	token.Header["kid"] = "e1bd3f64-6bc4-42b9-a294-1e4590959eb3"

	var keys jose.JSONWebKeySet
	cmdx.Must(json.NewDecoder(bytes.NewBufferString(key)).Decode(&keys), "")

	signed, err := token.SignedString(keys.Key("01cf6f34-07a8-40b7-b536-5fede7e2323e")[0].Key.(*rsa.PrivateKey))
	cmdx.Must(err, "%s", err)
	return signed
}()

func main() {
	res, body := requestWithJWT(tokenValid)
	if res.StatusCode != 200 {
		panic("proxy: expected 200: " + body)
	}

	res, body = requestWithJWT("not.valid.token")
	if res.StatusCode != 401 {
		panic("proxy: expected 401: " + body)
	}

	res, body = decisionWithJWT(tokenValid)
	if res.StatusCode != 200 {
		panic("decision: expected 200: " + body)
	}

	res, body = decisionWithJWT("not.valid.token")
	if res.StatusCode != 401 {
		panic("decision: expected 401: " + body)
	}
}

func requestWithJWT(token string) (*http.Response, string) {
	pu := x.ParseURLOrPanic(os.Getenv("OATHKEEPER_PROXY"))
	req, err := http.NewRequest("GET", urlx.AppendPaths(pu, "/jwt").String(), nil)
	cmdx.Must(err, "%s", err)
	req.Header.Set("Authorization", "Bearer "+token)
	res, err := http.DefaultClient.Do(req)
	cmdx.Must(err, "%s", err)
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	cmdx.Must(err, "%s", err)
	return res, string(body)
}

func decisionWithJWT(token string) (*http.Response, string) {
	pu := x.ParseURLOrPanic(os.Getenv("OATHKEEPER_API"))
	req, err := http.NewRequest("GET", urlx.AppendPaths(pu, "decisions", "jwt").String(), nil)
	cmdx.Must(err, "%s", err)
	req.Header.Set("Authorization", "Bearer "+token)
	res, err := http.DefaultClient.Do(req)
	cmdx.Must(err, "%s", err)
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	cmdx.Must(err, "%s", err)
	return res, string(body)
}
