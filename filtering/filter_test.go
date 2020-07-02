package filtering

import (
	"testing"

	ct "github.com/dfuse-io/dfuse-eosio/codec/testing"
	pbcodec "github.com/dfuse-io/dfuse-eosio/pb/dfuse/eosio/codec/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilterOut(t *testing.T) {
	tests := []struct {
		name         string
		include      string
		exclude      string
		trace        *pbcodec.TransactionTrace
		expectedPass bool
	}{
		{
			"filter nothing",
			"",
			"",
			ct.TrxTrace(t, ct.ActionTrace(t, "whatever:action")),
			true,
		},
		{
			"filter nothing, with default programs",
			"true",
			"false",
			ct.TrxTrace(t, ct.ActionTrace(t, "whatever:action")),
			true,
		},
		{
			"blacklist things FROM badguy",
			`true`,
			`account == "eosio.token" && data.from == "badguy"`,
			ct.TrxTrace(t, ct.ActionTrace(t, "eosio.token:transfer", ct.ActionData(`{"from":"goodguy","to":"badguy"}`))),
			true,
		},
		{
			"blacklist things TO badguy",
			`true`,
			"account == 'eosio.token' && data.to == 'badguy'",
			ct.TrxTrace(t, ct.ActionTrace(t, "eosio.token:transfer", ct.ActionData(`{"from":"goodguy","to":"badguy"}`))),
			false,
		},
		{
			"blacklist transfers to eidosonecoin",
			"*",
			`account == 'eidosonecoin' || receiver == 'eidosonecoin' || (account == 'eosio.token' && (data.to == 'eidosonecoin' || data.from == 'eidosonecoin'))`,
			ct.TrxTrace(t, ct.ActionTrace(t, "eosio.token:transfer", ct.ActionData(`{"from":"goodguy","to":"eidosonecoin"}`))),
			false,
		},
		{
			"non-matching identifier in filter-out program doesn't blacklist",
			"",
			`account == 'eosio.token' && data.from == 'broken'`,
			ct.TrxTrace(t, ct.ActionTrace(t, "eosio.token:issue", ct.ActionData(`{"to":"winner"}`))),
			true,
		},
		{
			"non-matching identifier in filter-on program still matches",
			`account == 'eosio.token' && data.bob == 'broken'`,
			``,
			ct.TrxTrace(t, ct.ActionTrace(t, "eosio.token:issue", ct.ActionData(`{"to":"winner"}`))),
			false,
		},
		{
			"both whitelist and blacklist fail",
			`data.bob == 'broken'`,
			`data.rita == 'rebroken'`,
			ct.TrxTrace(t, ct.ActionTrace(t, "any:any", ct.ActionData(`{"denise":"winner"}`))),
			false,
		},
		{
			"whitelisted but blacklist cleans out",
			`data.bob == '1'`,
			`data.rita == '2'`,
			ct.TrxTrace(t, ct.ActionTrace(t, "any:any", ct.ActionData(`{"bob":"1","rita":"2"}`))),
			false,
		},
		{
			"whitelisted but blacklist broken so doesn't clean out",
			`data.bob == '1'`,
			`data.broken == 'really'`,
			ct.TrxTrace(t, ct.ActionTrace(t, "any:any", ct.ActionData(`{"bob":"1"}`))),
			true,
		},

		{
			"block receiver",
			"",
			`receiver == "badguy"`,
			ct.TrxTrace(t, ct.ActionTrace(t, "badguy:any:any", ct.ActionData(`{}`))),
			false,
		},
		{
			"prevent a failure on evaluation, so matches because blacklist fails",
			"",
			`account == "badacct" && has(data.from) && data.from != "badguy"`,
			ct.TrxTrace(t, ct.ActionTrace(t, "badrecv:badacct:any", ct.ActionData(`{}`))),
			true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Len(t, test.trace.ActionTraces, 1, "This test accepts a single action trace per transaction trace currently")

			filter, err := NewBlockFilter(test.include, test.exclude)
			require.NoError(t, err)

			assert.Equal(t, test.expectedPass, filter.shouldProcess(test.trace, test.trace.ActionTraces[0]))
		})
	}
}

func TestCompileCELPrograms(t *testing.T) {
	_, err := NewBlockFilter("bro = '", "")
	require.Error(t, err)

	_, err = NewBlockFilter("", "ken")
	require.Error(t, err)
}
