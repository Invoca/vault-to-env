package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type arrayFlag []string

func (a *arrayFlag) Set(value string) error {
	*a = append(*a, strings.TrimSpace(value))
	return nil
}

func (a *arrayFlag) String() string {
	return fmt.Sprint([]string(*a))
}

func setEnvs(envVars, values []string) {
	for i, envVar := range envVars {
		os.Setenv(envVar, values[i])
	}
}

type vaultResponse struct {
	RequestID     string                 `json:"request_id"`
	LeaseID       string                 `json:"lease_id"`
	LeaseDuration int                    `json:"lease_duration"`
	Data          map[string]interface{} `json:"data"`
	Warnings      []string               `json:"warnings"`
	WrapInfo      string                 `json:"wrap_info,omitempty"`
	Auth          string                 `json:"auth,omitempty"`
}

func queryVault(url, token string, paths []string) []string {
	var client http.Client
	var values []string

	for _, path := range paths {
		vr := vaultResponse{}

		reqURL := url + "/v1/" + path
		req, err := http.NewRequest("GET", reqURL, nil)
		if err != nil {
			fmt.Printf("unable to create request; %v", err)
		}
		req.Header.Set("X-Vault-Token", token)

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("unable to get response; %v", err)
		}
		defer resp.Body.Close()

		if err := json.NewDecoder(resp.Body).Decode(&vr); err != nil {
			fmt.Printf("unable to decode response body; %v", err)
		}

		vrData, err := json.Marshal(&vr.Data)
		if err != nil {
			fmt.Printf("unable to marshal vrData; %v", err)
		}
		values = append(values, string(vrData))
	}

	return values
}

func parseKeys(keys, data []string) []string {
	var keyValue map[string]interface{}
	var values []string

	for i, key := range keys {
		if err := json.Unmarshal([]byte(data[i]), &keyValue); err != nil {
			fmt.Printf("unable to unmarshal keyValue; %v", err)
		}
		values = append(values, keyValue[key].(string))
	}

	return values
}

func main() {
	var paths, envVars, keys arrayFlag

	url := flag.String("url", "", "The Vault URL to query.")
	token := flag.String("token", "", "The token to query Vault with.")
	flag.Var(&paths, "path", "Path to secret being queried. Can be provided multiple times.")
	flag.Var(&envVars, "evar", "Env variable to store secret in. Can be provided multiple times.")
	flag.Var(&keys, "key", "Key to parse for the secret value")
	flag.Parse()

	results := queryVault(*url, *token, paths)
	keyValues := parseKeys(keys, results)
	setEnvs(envVars, keyValues)
}
