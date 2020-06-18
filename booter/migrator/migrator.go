package migrator

import (
	"fmt"

	bootops "github.com/dfuse-io/eosio-boot/ops"

	"go.uber.org/zap"

	"github.com/eoscanada/eos-go/ecc"

	"github.com/eoscanada/eos-go"

	rice "github.com/GeertJohan/go.rice"
	system "github.com/eoscanada/eos-go/system"
)

//go:generate rice embed-go

type Migrator struct {
	box         *rice.Box
	contract    eos.AccountName
	opPublicKey ecc.PublicKey
	actionChan  chan interface{}
	dataDir     string
}

func newMigrator(opPublicKey ecc.PublicKey, dataDir string, actionChan chan interface{}) *Migrator {
	return &Migrator{
		dataDir:     dataDir,
		opPublicKey: opPublicKey,
		box:         rice.MustFindBox("./code/build"),
		actionChan:  actionChan,
		contract:    eos.AN("dfuse.mgrt"),
	}
}

func (m *Migrator) newAccountActions(publicKey ecc.PublicKey, in chan interface{}) (err error) {
	in <- system.NewNewAccount("eosio", m.contract, publicKey)
	in <- system.NewBuyRAMBytes("eosio", m.contract, 100000)
	return
}

func (m *Migrator) setContractActions(contract eos.AccountName, in chan interface{}) error {
	abiCnt, err := readBoxFile(m.box, "migrator.abi")
	if err != nil {
		return fmt.Errorf("unable to open migration abi cnt: %w", err)
	}

	wasmCnt, err := readBoxFile(m.box, "migrator.wasm")
	if err != nil {
		return fmt.Errorf("unable to open migration wasm cnt: %w", err)
	}

	actions, err := system.NewSetContractContent(contract, wasmCnt, abiCnt)
	if err != nil {
		return fmt.Errorf("unable to create set contract actions: %w", err)
	}

	for _, action := range actions {
		in <- action
	}

	return nil
}

func (m *Migrator) migrateAccount(accountData *AccountData) error {
	zlog.Debug("processing account", zap.String("account", accountData.name))

	zlog.Debug("setting migrator code", zap.String("contract", accountData.name))
	err := m.setContractActions(AN(accountData.name), m.actionChan)
	if err != nil {
		return fmt.Errorf("unable to set migrator code for account: %w", err)
	}
	m.actionChan <- bootops.EndTransaction(m.opPublicKey) // end transaction

	err = accountData.Migrate()
	if err != nil {
		return fmt.Errorf("unable to migrate account: %w", err)
	}

	return nil
}
