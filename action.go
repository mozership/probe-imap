package main

import (
	"errors"

	"github.com/hashicorp/go-hclog"
	"github.com/mozership/probe/sdk"
)

// Action runs the IMAP commands of a step against a mail server.
type Action struct {
	log hclog.Logger
}

// keeps are the kinds of guard the action keeps to, as action.yml declares
// them: it connects only to a server the run allows. It does not keep to
// read-only, as its commands may change a mailbox, so a run under
// --read-only refuses it unless --allow-action names it.
var keeps = []string{sdk.KindAllowHost}

// Run runs the commands in with, under no guard.
func (a *Action) Run(with map[string]any) (map[string]any, error) {
	ret, _, err := a.RunStep(sdk.Call{With: with})
	return ret, err
}

// RunStep runs the commands in call.With, connecting only to a server the
// guard of the run allows.
func (a *Action) RunStep(call sdk.Call) (map[string]any, map[string]any, error) {
	ret, err := a.run(call.With, call.Guard)
	return ret, nil, err
}

func (a *Action) run(with map[string]any, guard sdk.Guard) (map[string]any, error) {
	if a.log == nil {
		a.log = hclog.NewNullLogger()
	}
	if len(with) == 0 {
		return map[string]any{}, errors.New("imap action requires parameters in 'with' section. Please specify connection details like host, username, password")
	}

	sdk.LogParams(a.log, "received imap request parameters", with)

	// The request and the response are logged as the parameters are, with
	// the password hidden and long values such as message bodies truncated.
	before := WithBefore(func(req *Req) {
		a.logStruct("imap request prepared", req)
	})
	after := WithAfter(func(res *Res) {
		a.logStruct("imap response received", res)
	})
	check := WithCheck(func(req *Req) error {
		if len(guard.AllowHosts) == 0 {
			return nil
		}
		return guard.CheckHost(req.Address())
	})
	ret, err := Request(with, check, before, after)

	sdk.LogOutcome(a.log, "imap request", ret, err)

	return ret, err
}

// logStruct records v, a Req or a Res, by its map tags through
// sdk.LogParams, which hides credentials such as the password.
func (a *Action) logStruct(msg string, v any) {
	m, err := sdk.EncodeResult(v)
	if err != nil {
		a.log.Trace(msg, "error", err)
		return
	}
	sdk.LogParams(a.log, msg, m)
}
