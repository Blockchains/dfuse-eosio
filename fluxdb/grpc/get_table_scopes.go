package grpc

import (
	"sort"

	"github.com/dfuse-io/derr"
	"github.com/dfuse-io/dfuse-eosio/fluxdb"
	pbfluxdb "github.com/dfuse-io/dfuse-eosio/pb/dfuse/eosio/fluxdb/v1"
	"github.com/dfuse-io/logging"
	"github.com/eoscanada/eos-go"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
)

func (s *Server) GetTableScopes(request *pbfluxdb.GetTableScopesRequest, stream pbfluxdb.FluxDB_GetTableScopesServer) error {
	ctx := stream.Context()
	zlogger := logging.Logger(ctx, zlog)
	zlogger.Debug("get table scopes",
		zap.Reflect("request", request),
	)

	blockNum := uint32(request.BlockNum)
	contract := eos.AccountName(request.Contract)
	table := eos.TableName(request.Table)
	actualBlockNum, _, _, speculativeWrites, err := s.prepareRead(ctx, blockNum, false)
	if err != nil {
		return derr.Statusf(codes.Internal, "unable to prepare read: %s", err)
	}

	tabletRows, err := s.db.ReadTabletAt(
		ctx,
		actualBlockNum,
		fluxdb.NewContractTableScopeTablet(request.Contract, request.Table),
		speculativeWrites,
	)
	if err != nil {
		return derr.Statusf(codes.Internal, "uanble to read tablet at %d: %s", blockNum, err)
	}

	zlogger.Debug("post-processing table scopes", zap.Int("table_scope_count", len(tabletRows)))
	scopes := sortedScopes(tabletRows)
	if len(scopes) == 0 {
		zlogger.Debug("no scopes found for request, checking if we ever see this table")
		seen, err := s.db.HasSeenTableOnce(ctx, contract, table)
		if err != nil {
			return derr.Statusf(codes.Internal, "unable to know if table was seen once in db: %s", err)
		}

		if !seen {
			return derr.Statusf(codes.Internal, "table %s/%s does not exist in ABI at this block height", request.Contract, request.Table)
		}
	}

	// TODO: pass block num in header
	for _, scope := range scopes {
		stream.Send(&pbfluxdb.TableScopeResponse{
			BlockNum: uint64(actualBlockNum),
			Scope:    string(scope),
		})
	}

	return nil
}

func sortedScopes(tabletRows []fluxdb.TabletRow) (out []string) {
	if len(tabletRows) <= 0 {
		return
	}

	out = make([]string, len(tabletRows))
	for i, tabletRow := range tabletRows {
		out[i] = tabletRow.(*fluxdb.ContractTableScopeRow).Scope()
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i] < out[j]
	})

	return
}
