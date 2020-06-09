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

// Contract State
const cstPrefix = "cst"

func init() {
	RegisterTabletFactory(cstPrefix, func(row *pbfluxdb.TabletRow) Tablet {
		return ContractStateTablet(fmt.Sprintf("%s/%s", cstPrefix, row.TabletKey))
	})
}

func NewContractStateTablet(contract, scope, table string) ContractStateTablet {
	return ContractStateTablet(fmt.Sprintf("%s/%s:%s:%s", cstPrefix, contract, scope, table))
}

type ContractStateTablet string

func (t ContractStateTablet) Key() string {
	return string(t)
}

func (t ContractStateTablet) Explode() (collection, contract, scope, table string) {
	segments := strings.Split(string(t), "/")
	tabletParts := strings.Split(segments[1], ":")

	return segments[0], tabletParts[0], tabletParts[1], tabletParts[2]
}

func (t ContractStateTablet) KeyForRowAt(blockNum uint32, primaryKey string) string {
	return t.KeyAt(blockNum) + "/" + string(primaryKey)
}

func (t ContractStateTablet) KeyAt(blockNum uint32) string {
	return string(t) + "/" + HexBlockNum(blockNum)
}

func (t ContractStateTablet) NewRow(blockNum uint32, primaryKey string, payer string, data []byte, isDeletion bool) *ContractStateRow {
	row := &ContractStateRow{
		BaseTabletRow: BaseTabletRow{pbfluxdb.TabletRow{
			Collection:  cstPrefix,
			TabletKey:   t.Key(),
			BlockNumKey: HexBlockNum(blockNum),
			PrimKey:     primaryKey,
		}},
	}

	if !isDeletion {
		row.Payload = make([]byte, len(data)+8)
		binary.BigEndian.PutUint64(row.Payload, NA(eos.Name(payer)))
		copy(row.Payload[8:], data)
	}

	return row
}

func (t ContractStateTablet) NewRowFromKV(key string, value []byte) (TabletRow, error) {
	if len(value) < 8 {
		return nil, errors.New("contract state tablet row value should have at least 8 bytes (payer)")
	}

	_, tabletKey, blockNumKey, primaryKey, err := ExplodeTabletRowKey(key)
	if err != nil {
		return nil, fmt.Errorf("unable to explode tablet row key %q: %s", key, err)
	}

	return &ContractStateRow{
		BaseTabletRow: BaseTabletRow{pbfluxdb.TabletRow{
			Collection:  cstPrefix,
			TabletKey:   tabletKey,
			BlockNumKey: blockNumKey,
			PrimKey:     primaryKey,
			Payload:     value,
		}},
	}, nil
}

func (t ContractStateTablet) String() string {
	return string(t)
}

// IndexableTablet

func (t ContractStateTablet) PrimaryKeyByteCount() int {
	return 8
}

func (t ContractStateTablet) EncodePrimaryKey(buffer []byte, primaryKey string) error {
	binary.BigEndian.PutUint64(buffer, NA(eos.Name(primaryKey)))
	return nil
}

func (t ContractStateTablet) DecodePrimaryKey(buffer []byte) (primaryKey string, err error) {
	return eos.NameToString(binary.BigEndian.Uint64(buffer)), nil
}

// Row

type ContractStateRow struct {
	BaseTabletRow
}

func NewContractStateRow(blockNum uint32, op *pbcodec.DBOp) *ContractStateRow {
	tablet := NewContractStateTablet(op.Code, op.Scope, op.TableName)
	isDeletion := op.Operation == pbcodec.DBOp_OPERATION_REMOVE

	var payer string
	var data []byte
	if !isDeletion {
		payer = op.NewPayer
		data = op.NewData
	}

	return tablet.NewRow(blockNum, op.PrimaryKey, payer, data, isDeletion)
}

func (r *ContractStateRow) Payer() string {
	return eos.NameToString(binary.BigEndian.Uint64(r.Payload))
}

func (r *ContractStateRow) Data() []byte {
	return r.Payload[8:]
}
