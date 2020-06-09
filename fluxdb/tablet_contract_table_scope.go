package fluxdb

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strings"

	pbcodec "github.com/dfuse-io/dfuse-eosio/pb/dfuse/eosio/codec/v1"
	pbfluxdb "github.com/dfuse-io/dfuse-eosio/pb/dfuse/eosio/fluxdb/v1"
	"github.com/eoscanada/eos-go"
)

// Contract Table Scope
const ctblsPrefix = "ctbls"

func init() {
	RegisterTabletFactory(ctblsPrefix, func(row *pbfluxdb.TabletRow) Tablet {
		return ContractTableScopeTablet(fmt.Sprintf("%s/%s", ctblsPrefix, row.TabletKey))
	})
}

func NewContractTableScopeTablet(contract, table string) ContractTableScopeTablet {
	return ContractTableScopeTablet(fmt.Sprintf("%s/%s:%s", ctblsPrefix, contract, table))
}

type ContractTableScopeTablet string

func (t ContractTableScopeTablet) Key() string {
	return string(t)
}

func (t ContractTableScopeTablet) Explode() (collection, contract, table string) {
	segments := strings.Split(string(t), "/")
	tabletParts := strings.Split(segments[1], ":")

	return segments[0], tabletParts[0], tabletParts[1]
}

func (t ContractTableScopeTablet) KeyForRowAt(blockNum uint32, primaryKey string) string {
	return t.KeyAt(blockNum) + "/" + string(primaryKey)
}

func (t ContractTableScopeTablet) KeyAt(blockNum uint32) string {
	return string(t) + "/" + HexBlockNum(blockNum)
}

func (t ContractTableScopeTablet) NewRow(blockNum uint32, scope string, payer string, isDeletion bool) *ContractTableScopeRow {
	row := &ContractTableScopeRow{
		BaseTabletRow: BaseTabletRow{pbfluxdb.TabletRow{
			Collection:  ctblsPrefix,
			TabletKey:   t.Key(),
			BlockNumKey: HexBlockNum(blockNum),
			PrimKey:     scope,
		}},
	}

	if !isDeletion {
		row.Payload = make([]byte, 8)
		binary.BigEndian.PutUint64(row.Payload, NA(eos.Name(payer)))
	}

	return row
}

func (t ContractTableScopeTablet) NewRowFromKV(key string, value []byte) (TabletRow, error) {
	if len(value) == 0 || len(value) == 8 {
		return nil, errors.New("contract table scope row value should have at 0 byte (deletion) or 8 bytes (payer)")
	}

	_, tabletKey, blockNumKey, primaryKey, err := ExplodeTabletRowKey(key)
	if err != nil {
		return nil, fmt.Errorf("unable to explode tablet row key %q: %s", key, err)
	}

	return &ContractTableScopeRow{
		BaseTabletRow: BaseTabletRow{pbfluxdb.TabletRow{
			Collection:  ctblsPrefix,
			TabletKey:   tabletKey,
			BlockNumKey: blockNumKey,
			PrimKey:     primaryKey,
			Payload:     value,
		}},
	}, nil
}

func (t ContractTableScopeTablet) String() string {
	return string(t)
}

func (t ContractTableScopeTablet) PrimaryKeyByteCount() int {
	return 8
}

func (t ContractTableScopeTablet) EncodePrimaryKey(buffer []byte, primaryKey string) error {
	binary.BigEndian.PutUint64(buffer, NA(eos.Name(primaryKey)))
	return nil
}

func (t ContractTableScopeTablet) DecodePrimaryKey(buffer []byte) (primaryKey string, err error) {
	return eos.NameToString(binary.BigEndian.Uint64(buffer)), nil
}

type ContractTableScopeRow struct {
	BaseTabletRow
}

func NewContractTableScopeRow(blockNum uint32, op *pbcodec.TableOp) *ContractTableScopeRow {
	tablet := NewContractTableScopeTablet(op.Code, op.TableName)
	isDeletion := op.Operation == pbcodec.TableOp_OPERATION_REMOVE

	var payer string
	if !isDeletion {
		payer = op.Payer
	}

	return tablet.NewRow(blockNum, op.Scope, payer, isDeletion)
}

func (r *ContractTableScopeRow) Payer() string {
	return eos.NameToString(binary.BigEndian.Uint64(r.Payload))
}
