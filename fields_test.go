package main

import (
	"os"
	"slices"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/mozership/probe/sdk"
)

// TestManifestsDeclare checks that action.yml, and the one the e2e workflow
// uses, declare the params, the fields of res and the guard the action has,
// so that probe check and the guard of a run take it as it is.
func TestManifestsDeclare(t *testing.T) {
	for _, path := range []string{"action.yml", "e2e/action/action.yml"} {
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		m, err := sdk.ParseManifest(b)
		if err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		if want := sdk.ParamNames(Req{}); !slices.Equal(m.Params, want) {
			t.Errorf("%s declares params %v, want %v", path, m.Params, want)
		}
		if want := sdk.ParamNames(Res{}); !slices.Equal(m.Res, want) {
			t.Errorf("%s declares res %v, want %v", path, m.Res, want)
		}
		if !slices.Equal(m.Guard, keeps) {
			t.Errorf("%s declares guard %v, want %v", path, m.Guard, keeps)
		}
	}
}

// TestResultDeclared runs commands of each kind and checks that the result
// has no field of res the manifest leaves out.
func TestResultDeclared(t *testing.T) {
	host, port := plainIMAPServer(t, 2)
	a := &Action{log: hclog.NewNullLogger()}
	ret, err := a.Run(map[string]any{
		"host": host, "port": port, "username": "user", "password": "pass", "tls": false, "timeout": "5s",
		"commands": []any{
			map[string]any{"name": "select", "mailbox": "INBOX"},
			map[string]any{"name": "search", "criteria": map[string]any{"flags": []any{}}},
			map[string]any{"name": "fetch", "sequence": "1:2", "dataitem": "FLAGS"},
			map[string]any{"name": "noop"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	res, ok := ret["res"].(map[string]any)
	if !ok {
		t.Fatalf("res = %T", ret["res"])
	}
	declared := sdk.ParamNames(Res{})
	for k := range res {
		if !slices.Contains(declared, k) {
			t.Errorf("res.%s is not declared", k)
		}
	}
	for _, k := range []string{"req", "res", "rt", "status"} {
		if _, ok := ret[k]; !ok {
			t.Errorf("the result has no %s", k)
		}
	}
}

func TestRequired(t *testing.T) {
	a := &Action{log: hclog.NewNullLogger()}
	with := map[string]any{"host": "127.0.0.1", "port": 1, "username": "user", "password": "pass", "tls": false, "timeout": "1s"}
	for _, key := range []string{"host", "username", "password"} {
		w := map[string]any{}
		for k, v := range with {
			if k != key {
				w[k] = v
			}
		}
		if _, err := a.Run(w); err == nil {
			t.Errorf("Run() without %s succeeded", key)
		}
	}
}
