package main

import (
	"strings"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/mozership/probe/sdk"
)

func TestRunStep_allowHost(t *testing.T) {
	host, port := plainIMAPServer(t, 1)
	with := map[string]any{
		"host": host, "port": port, "username": "user", "password": "pass", "tls": false, "timeout": "5s",
		"commands": []any{map[string]any{"name": "select", "mailbox": "INBOX"}},
	}
	a := &Action{log: hclog.NewNullLogger()}

	ret, _, err := a.RunStep(sdk.Call{With: with, Guard: sdk.Guard{AllowHosts: []string{host}}})
	if err != nil {
		t.Fatalf("RunStep() on an allowed host = %v", err)
	}
	if ret["status"] != 0 {
		t.Errorf("RunStep() = %v", ret)
	}

	_, _, err = a.RunStep(sdk.Call{With: with, Guard: sdk.Guard{AllowHosts: []string{"mail.example.com"}}})
	if !sdk.IsRefused(err) || !strings.Contains(err.Error(), "is not one the run allows") {
		t.Fatalf("RunStep() on a host not allowed = %v, want a refusal", err)
	}
}

func TestRunStep_allowHostDefaultPort(t *testing.T) {
	a := &Action{log: hclog.NewNullLogger()}
	with := map[string]any{"host": "imap.example.com", "username": "user", "password": "pass", "timeout": "1s"}
	_, _, err := a.RunStep(sdk.Call{With: with, Guard: sdk.Guard{AllowHosts: []string{"imap.example.com:143"}}})
	if !sdk.IsRefused(err) || !strings.Contains(err.Error(), "imap.example.com:993") {
		t.Fatalf("RunStep() = %v, want a refusal of port 993", err)
	}
}
