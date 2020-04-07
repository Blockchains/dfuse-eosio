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

package server

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/dfuse-io/derr"
	eos "github.com/eoscanada/eos-go"
	"github.com/dfuse-io/logging"
	"github.com/dfuse-io/validator"
	"go.uber.org/zap"
)

func (srv *EOSServer) getTableRowHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	zlog := logging.Logger(ctx, zlog)

	errors := validateGetTableRowRequest(r)
	if len(errors) > 0 {
		writeError(ctx, w, derr.RequestValidationError(ctx, errors))
		return
	}

	request := extractGetTableRowRequest(r)
	zlog.Debug("extracted request", zap.Reflect("request", request))

	actualBlockNum, lastWrittenBlockID, upToBlockID, speculativeWrites, err := srv.prepareRead(ctx, request.BlockNum, request.IrreversibleOnly)
	if err != nil {
		writeError(ctx, w, derr.Wrap(err, "prepare read failed"))
		return
	}

	tableRowResponse, err := srv.readTableRow(
		ctx,
		actualBlockNum,
		request.Account,
		request.Table,
		request.Scope,
		request.PrimaryKey,
		request.readRequestCommon,
		getKeyConverterForType(request.KeyType),
		speculativeWrites,
	)

	if err != nil {
		writeError(ctx, w, derr.Wrap(err, "read table row failed"))
		return
	}

	response := &getTableRowResponse{
		commonStateResponse: newCommonGetResponse(upToBlockID, lastWrittenBlockID),
		Row:                 tableRowResponse.Row,
	}

	zlog.Debug("streaming response", zap.Reflect("common_response", response.commonStateResponse))
	streamResponse(ctx, w, response)
}

type getTableRowRequest struct {
	*readRequestCommon

	IrreversibleOnly bool   `json:"irreversible_only"`
	Account          string `json:"account"`
	Table            string `json:"table"`
	Scope            string `json:"scope"`
	PrimaryKey       string `json:"primary_key"`
}

type getTableRowResponse struct {
	*commonStateResponse
	ABI *eos.ABI  `json:"abi,omitempty"`
	Row *tableRow `json:"row,omitempty"`
}

func validateGetTableRowRequest(r *http.Request) url.Values {
	errors := validator.ValidateQueryParams(r, withCommonValidationRules(validator.Rules{
		"account":           []string{"required", "fluxdb.eos.name"},
		"table":             []string{"required", "fluxdb.eos.name"},
		"scope":             []string{"fluxdb.eos.extendedName"},
		"primary_key":       []string{"required"},
		"irreversible_only": []string{"bool"},
	}))

	// Let's ensure the scope param is at least present (but can be the empty string)
	if _, ok := r.Form["scope"]; !ok {
		errors["scope"] = []string{"The scope field is required"}
	}

	// FIXME (MATT): Deal with KeyType here and conversion from key to uint64 to check if conversion if correct

	return errors
}

func extractGetTableRowRequest(r *http.Request) *getTableRowRequest {
	irreversibleOnly, _ := strconv.ParseBool(r.FormValue("irreversible_only"))

	// FIXME (MATT): Convert from string key to actual uint64 then format the uint64 as %016x

	return &getTableRowRequest{
		readRequestCommon: extractReadRequestCommon(r),

		Table:            r.FormValue("table"),
		Account:          r.FormValue("account"),
		Scope:            r.FormValue("scope"),
		PrimaryKey:       r.FormValue("primary_key"),
		IrreversibleOnly: irreversibleOnly,
	}
}
