package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
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

func split(kv string) (string, string) {
	kva := strings.Split(kv, "=")
	return kva[0], kva[1]
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

func (r *vaultResponse) buildExports(eks []string) ([]string, error) {
	var values []string

	if len(eks) != len(r.Data) {
		return nil, fmt.Errorf("You must provide the same amount of eks (%d) as values in your secret (%d)", len(eks), len(r.Data))
	}
	for _, ek := range eks {
		e, k := split(ek)

		values = append(values, fmt.Sprintf("export %s=%s", e, r.Data[k]))
	}

	return values, nil
}

func queryVault(url, token, path string) (vaultResponse, error) {
	var client http.Client

	vr := vaultResponse{}

	reqURL := url + "/v1/" + path
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return vr, fmt.Errorf("Unable to create request; %v", err)
	}
	req.Header.Set("X-Vault-Token", token)

	resp, err := client.Do(req)
	if err != nil {
		return vr, fmt.Errorf("Unable to get response; %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return vr, fmt.Errorf("Did not get back %v, got; %v", http.StatusOK, resp.StatusCode)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&vr); err != nil {
		return vr, fmt.Errorf("Unable to decode response body; %v", err)
	}

	return vr, nil
}

func main() {
	var eks arrayFlag

	url := flag.String("url", "", "The Vault URL to query.")
	token := flag.String("token", "", "The token to query Vault with.")
	path := flag.String("path", "", "Path to secret being queried.")
	flag.Var(&eks, "eks", "ENV=key pairing where ENV gets set to the value of key in Vault")
	flag.Parse()

	response, err := queryVault(*url, *token, *path)
	if err != nil {
		log.Fatal(err)
	}
	keyValues, _ := response.buildExports(eks)
	fmt.Printf(strings.Join(keyValues, "\n"))
}
