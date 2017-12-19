package main

import (
	. "flag"
	"fmt"
	"net/http"
	"reflect"
	"strings"
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
	url := "http://localhost:8200"
	token := "roottoken"
	path := "secret/password"
	responseBody := `{"data":{"value1":"foo"}}`
	expected := "foo"

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		fmt.Sprintf("http://localhost:8200/v1/%s", path),
		httpmock.NewStringResponder(http.StatusOK, responseBody),
	)

	results, _ := queryVault(url, token, path)
	if results.Data["value1"] != expected {
		t.Errorf("expected value %v; got %v", expected, results.Data["value1"])
	}
}

func TestQueryVaultErr(t *testing.T) {
	tt := []struct {
		url         string
		token       string
		path        string
		mock        bool
		mStatusCode int
		mBody       string
		errorMsg    string
	}{
		{
			url:      "http://not a.url/",
			errorMsg: "Unable to create request",
		}, {
			url:      "",
			errorMsg: "Unable to get response",
		}, {
			url:         "http://localhost:8200",
			mock:        true,
			mStatusCode: http.StatusBadRequest,
			errorMsg:    "Did not get back",
		}, {
			url:         "http://localhost:8200",
			mock:        true,
			mStatusCode: http.StatusOK,
			mBody:       "this is not json",
			errorMsg:    "Unable to decode response body",
		},
	}
	for _, test := range tt {
		if test.mock {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()

			httpmock.RegisterResponder(
				"GET",
				fmt.Sprintf("%s/v1/%s", test.url, test.path),
				httpmock.NewStringResponder(test.mStatusCode, test.mBody),
			)
		}
		_, err := queryVault(test.url, test.token, test.path)
		if err == nil || !strings.Contains(err.Error(), test.errorMsg) {
			t.Errorf("expected; %v, got; %v", test.errorMsg, err)
		}
	}
}

func TestBuildExports(t *testing.T) {
	tt := []struct {
		desc        string
		vr          vaultResponse
		eks         []string
		expected    []string
		expectError bool
	}{
		{
			desc: "equal lengths",
			vr: vaultResponse{Data: map[string]interface{}{
				"value": "itsasecret",
				"biz":   "bar",
			}},
			eks:      []string{"SECRET=value", "FOO=biz"},
			expected: []string{"export SECRET=itsasecret", "export FOO=bar"},
		}, {
			desc: "unequal lengths",
			vr: vaultResponse{Data: map[string]interface{}{
				"value": "itsasecret",
				"biz":   "bar",
			}},
			eks:         []string{"SECRET=value"},
			expectError: true,
		},
	}

	for _, test := range tt {
		t.Run(test.desc, func(t *testing.T) {
			results, err := test.vr.buildExports(test.eks)

			if test.expectError && err == nil {
				t.Errorf("expected error")
			}

			if !reflect.DeepEqual(results, test.expected) {
				t.Errorf("expected %v; got %v", test.expected, results)
			}
		})
	}
}
