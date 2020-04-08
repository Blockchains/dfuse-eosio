package cli

import (
	pbbstream "github.com/dfuse-io/pbgo/dfuse/bstream/v1"
)

const (
	Protocol                    pbbstream.Protocol = pbbstream.Protocol_EOS
	DmeshServiceVersion         string             = "v1"
	DmeshNamespace              string             = "local"
	NetworkID                   string             = "eos-local"
	EosManagerHTTPAddr          string             = ":13008"
	EosMindreaderHTTPAddr       string             = ":13009"
	MindreaderGRPCAddr          string             = ":13010"
	RelayerServingAddr          string             = ":13011"
	MergerServingAddr           string             = ":13012"
	AbiServingAddr              string             = ":13013"
	BlockmetaServingAddr        string             = ":13014"
	ArchiveServingAddr          string             = ":13015"
	ArchiveHTTPServingAddr      string             = ":13016"
	LiveServingAddr             string             = ":13017"
	RouterServingAddr           string             = ":13018"
	RouterHTTPServingAddr       string             = ":13019"
	KvdbHTTPServingAddr         string             = ":13020"
	IndexerServingAddr          string             = ":13021"
	IndexerHTTPServingAddr      string             = ":13022"
	TokenmetaServingAddr        string             = ":13023" // Not implemented yet, present for booting purposes, does not work
	DgraphqlHTTPServingAddr     string             = ":13024"
	DgraphqlGrpcServingAddr     string             = ":13025"
	DashboardGrpcServingAddr    string             = ":13726"
	EoswsHTTPServingAddr        string             = ":13027"
	ForkresolverServingAddr     string             = ":13028"
	ForkresolverHTTPServingAddr string             = ":13029"
	DashboardHTTPListenAddr     string             = ":8080"
	EosqHTTPServingAddress      string             = ":8081"
)
