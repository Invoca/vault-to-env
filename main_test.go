package main

import (
	. "flag"
	"net/http"
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
	tt := []struct {
		Desc         string
		URL          string
		Token        string
		Path         string
		ResponseBody string
		ResponseCode int
		Expected     string
		TestError    bool
		Error        string
	}{
		{
			Desc:         "Normal case",
			URL:          "http://localhost:8200",
			Token:        "token",
			Path:         "secret/password",
			ResponseBody: `{"data":{"value1":"foo"}}`,
			ResponseCode: http.StatusOK,
			Expected:     "foo",
		}, {
			Desc:      "Unable to get response",
			URL:       "http://localhost:8200",
			TestError: true,
		}, {
			Desc:      "Bad URL",
			URL:       "notaurl",
			TestError: true,
		},
	}

	for _, test := range tt {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder(
			"GET",
			"http://localhost:8200/v1/secret/password",
			httpmock.NewStringResponder(http.StatusOK, test.ResponseBody),
		)

		results, err := queryVault(test.URL, test.Token, test.Path)
		if err != nil && test.TestError {
			// pass
		} else if results.Data["value1"] != test.Expected {
			t.Errorf("expected value %v; got %v", test.Expected, results.Data["value1"])
		}
	}
}

func TestBuildExports(t *testing.T) {
	eks := []string{"SECRET=value", "FOO=biz"}
	vr := vaultResponse{Data: map[string]interface{}{
		"value": "itsasecret",
		"biz":   "bar",
	}}

	expected := []string{"export SECRET=itsasecret", "export FOO=bar"}
	results, _ := vr.buildExports(eks)

	if !reflect.DeepEqual(results, expected) {
		t.Errorf("expected %v; got %v", expected, results)
	}
}
