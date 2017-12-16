package main

import (
	. "flag"
	"net/http"
	"os"
	"reflect"
	"testing"

	"github.com/jarcoal/httpmock"
)

func TestArrayFlagSet(t *testing.T) {
	var flags FlagSet
	flags.Init("test", ContinueOnError)

	var a arrayFlag
	flags.Var(&a, "a", "usage")
	if err := flags.Parse([]string{"-a", "foo", "-a", "bar", "-a=foobar"}); err != nil {
		t.Error(err)
	}
	if len(a) != 3 {
		t.Fatal("expected 3 args; got ", len(a))
	}
	expect := "[foo bar foobar]"
	if a.String() != expect {
		t.Errorf("expected value %q got %q", expect, a.String())
	}
}

func TestSetEnvs(t *testing.T) {
	tt := []struct {
		Eks    []string
		Envs   []string
		Values []string
	}{
		{Eks: []string{"FOO=bar"}, Envs: []string{"FOO"}, Values: []string{"bar"}},
		{Eks: []string{"FOO=bar", "BIZ=baz"}, Envs: []string{"FOO", "BIZ"}, Values: []string{"bar", "baz"}},
	}
	for _, test := range tt {
		setEnvs(test.Eks)
		for i, env := range test.Envs {
			if result := os.Getenv(env); result != test.Values[i] {
				t.Errorf("expected value %v; got %v", test.Values[i], result)
			}
		}
	}
}

func TestSplit(t *testing.T) {
	kv := "a=b"
	key := "a"
	value := "b"

	k, v := split(kv)
	if k != key || v != value {
		t.Errorf("key should have been %v, was %v; value should have been %v, was %v", key, k, value, v)
	}
}

func TestQueryVault(t *testing.T) {
	url := "http://localhost:8200"
	token := "roottoken"
	path := "secret/password"
	responseJSON := `{"data":{"value1":"itsasecret", "value2":"noitsnot"}}`

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"http://localhost:8200/v1/secret/password",
		httpmock.NewStringResponder(http.StatusOK, responseJSON),
	)

	results := queryVault(url, token, path)

	if results.Data["value1"] != "itsasecret" {
		t.Errorf("expected value itsasecret; got %v", results.Data["value1"])
	}
	if results.Data["value2"] != "noitsnot" {
		t.Errorf("expected value noitsnot; got %v", results.Data["value2"])
	}
}

func TestParseKeys(t *testing.T) {
	eks := []string{"SECRET=value", "FOO=biz"}
	vr := vaultResponse{Data: map[string]interface{}{
		"value": "itsasecret",
		"biz":   "bar",
	}}

	expected := []string{"SECRET=itsasecret", "FOO=bar"}
	results, _ := vr.parseKeys(eks)

	if !reflect.DeepEqual(results, expected) {
		t.Errorf("expected %v; got %v", expected, results)
	}
}
