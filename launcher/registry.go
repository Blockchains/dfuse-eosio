package core

import (
	"github.com/dfuse-io/dfuse-eosio/metrics"
)

var AppRegistry = map[string]*AppDef{}

func RegisterApp(appDef *AppDef) {
	AppRegistry[appDef.ID] = appDef
}

func GetMetricAppMeta() map[string]*metrics.AppMeta {
	mapping := make(map[string]*metrics.AppMeta)
	for _, appDef := range AppRegistry {
		mapping[appDef.MetricsID] = &metrics.AppMeta{
			Title: appDef.Title,
			Id:    appDef.ID,
		}
	}
	return mapping
}
