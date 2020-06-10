// Copyright 2020 dfuse Platform Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fluxdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

//func TestAuthLink_BuildData(t *testing.T) {
//	updatedRow := &AuthLinkRow{
//		PermissionName: N(""),
//	}
//
//	updatedRowValue := updatedRow.buildData()
//
//	assert.Equal(t, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, updatedRowValue)
//}

func TestTabletRowKeys(t *testing.T) {
	tests := []struct {
		name     string
		row      TabletRow
		expected string
	}{
		{
			name:     "auth_link_tablet_row",
			row:      mustCreateAuthLinkTabletRow("eoscanadacom", 0, "token", "transfer", "active", false),
			expected: "al/eoscanadacom/00000000/token:transfer",
		},
		{
			name:     "key_account_tablet_row",
			row:      mustCreateKeyAccountTabletRow("EOS5MHPYyhjBjnQZejzZHqHewPWhGTfQWSVTWYEhDmJu4SXkzgweP", 0, "eosio", "active", false),
			expected: "ka/EOS5MHPYyhjBjnQZejzZHqHewPWhGTfQWSVTWYEhDmJu4SXkzgweP/00000000/eosio:active",
		},
		{
			name:     "contract_state_tablet_row",
			row:      mustCreateContractStateTabletRow("eosio", "scope", "table", 0, "key", "payer", nil, false),
			expected: "cst/eosio:scope:table/00000000/key",
		},
		{
			name:     "contract_tablet_scope_tablet_row",
			row:      mustCreateContractTableScopeTabletRow("eosio", "table", 0, "scope", "payer", false),
			expected: "ctbls/eosio:table/00000000/scope",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, test.row.Key())
		})
	}
}

//func TestWhoHasWhatShardContentiousTableNames(t *testing.T) {
//	include := func(tabletKey string, shardCount uint32) uint32 {
//		h := md5.New()
//		_, _ = h.Write([]byte(tabletKey))
//		md5Hash := h.Sum(nil)
//
//		bigInt := binary.LittleEndian.Uint32(md5Hash)
//
//		elementShard := bigInt % shardCount
//
//		return elementShard
//	}
//
//	// Belongs to shard 4, out of 100
//	row := NewContractStateTablet("eosio", "eosio", "global")
//	assert.Equal(t, 0, int(include(row.Key(), 100)))
//
//	row = NewContractStateTablet("eosio", "eosio", "voters")
//	assert.Equal(t, 52, int(include(row.Key(), 100)))
//
//	row = NewContractStateTablet("eosio", "eosbetdice11", "globalvars")
//	assert.Equal(t, 92, int(include(row.Key(), 100)))
//
//	// Belongs to shard ?
//	assert.Equal(t, 22, int(include("brl", 100)))
//}
