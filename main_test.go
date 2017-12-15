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
	envVars := []string{"foo", "bar"}
	values := []string{"biz", "baz"}

	setEnvs(envVars, values)

	for i, envVar := range envVars {
		if result := os.Getenv(envVar); values[i] != result {
			t.Errorf("expected %v; got %v", values[i], result)
		}
	}
}

func TestQueryVault(t *testing.T) {
	url := "http://localhost:8200"
	token := "roottoken"
	paths := []string{"secret/password"}
	values := []string{`{"value":"itsasecret"}`}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"http://localhost:8200/v1/secret/password",
		httpmock.NewStringResponder(http.StatusOK, `{"data":{"value":"itsasecret"}}`),
	)

	results := queryVault(url, token, paths)

	if !reflect.DeepEqual(results, values) {
		t.Errorf("expected %v; got %v", values, results)
	}
}

func TestParseKeys(t *testing.T) {
	keys := []string{"foo", "biz"}
	expected := []string{"bar", "zib"}
	data := []string{`{"foo":"bar"}`, `{"biz":"zib"}`}

	results := parseKeys(keys, data)

	if !reflect.DeepEqual(results, expected) {
		t.Errorf("expected %v; got %v", expected, results)
	}
}
