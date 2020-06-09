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
	"context"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/dfuse-io/derr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opencensus.io/trace"
)

func TestReadWithSpeculative(t *testing.T) {
	db, closer := NewTestDB(t)
	defer closer()

	blockNum := uint32(123)
	contract := "eosio"
	scope := "eoscanada"
	table := "delband"
	key := "...........1"

	writeBatchOfRequests(t, db, writeEmptyABI(t, blockNum, contract))

	contractStateTablet := NewContractStateTablet(contract, scope, table)
	speculativeWrites := []*WriteRequest{
		tabletRows(blockNum, contractStateTablet.NewRow(blockNum, key, "", nil, true)),
	}

	rows, err := db.ReadTabletAt(context.Background(), 123, contractStateTablet, speculativeWrites)

	require.NoError(t, err)
	require.Len(t, rows, 0)
}

func TestReadGetABI(t *testing.T) {
	acct := "eosio"
	traceID := fixedTraceID("00000000000000000000000000000001")
	spanContext := trace.SpanContext{TraceID: traceID}
	ctx, _ := trace.StartSpanWithRemoteParent(context.Background(), "test", spanContext)

	tests := []struct {
		name          string
		abis          []uint32
		fetchForBlock uint32
		expectedABI   string
		expectedError error
	}{
		{
			name: "fetch after last",
			abis: []uint32{
				3, 5,
			},
			fetchForBlock: 6,
			expectedABI:   `5`,
		},
		{
			name: "fetch between two",
			abis: []uint32{
				3, 5,
			},
			fetchForBlock: 4,
			expectedABI:   `3`,
		},
		{
			name: "fetch on the betweener",
			abis: []uint32{
				3, 4, 5,
			},
			fetchForBlock: 4,
			expectedABI:   `4`,
		},
		{
			name: "fetch on last",
			abis: []uint32{
				3, 5,
			},
			fetchForBlock: 5,
			expectedABI:   `5`,
		},
		{
			name: "fetch on first",
			abis: []uint32{
				3, 5,
			},
			fetchForBlock: 3,
			expectedABI:   `3`,
		},
		{
			name: "fetch before first",
			abis: []uint32{
				3, 5,
			},
			fetchForBlock: 2,
			expectedError: DataABINotFoundError(ctx, "eosio", 2),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db, closer := NewTestDB(t)
			defer closer()

			for _, abiBlock := range test.abis {
				writeBatchOfRequests(t, db, writePackedABI(t, abiBlock, acct, []byte(fmt.Sprintf("%d", abiBlock))))
			}

			siglet := NewContractABISiglet(acct)
			abiEntry, err := db.ReadSigletEntryAt(ctx, siglet, test.fetchForBlock, nil)
			if test.expectedError != nil {
				assertError(t, test.expectedError, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, abiEntry, "abi entry is nil")
				assert.Equal(t, test.expectedABI, string(abiEntry.(*ContractABIEntry).PackedABI()))
			}
		})
	}
}

func assertError(t *testing.T, expected error, actual error) {
	require.Error(t, actual)

	switch v := expected.(type) {
	case *derr.ErrorResponse:
		assertErrorResponse(t, v, actual)
	default:
		assert.Equal(t, expected, actual)
	}
}

func assertErrorResponse(t *testing.T, expected *derr.ErrorResponse, actual error) {
	v, ok := actual.(*derr.ErrorResponse)
	require.True(t, ok, "actual value must be a *derr.ErrorResponse type")

	assert.Equal(t, expected, v)
}

func fixedTraceID(hexInput string) (out trace.TraceID) {
	rawTraceID, _ := hex.DecodeString(hexInput)
	copy(out[:], rawTraceID)

	return
}
