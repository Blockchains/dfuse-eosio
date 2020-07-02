package cli

import (
	"github.com/dfuse-io/dlauncher/dashboard"
	"github.com/dfuse-io/dlauncher/launcher"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	launcher.RegisterApp(&launcher.AppDef{
		ID:          "dashboard",
		Title:       "Dashboard",
		Description: "dfuse for EOSIO - dashboard",
		MetricsID:   "dashboard",
		Logger:      launcher.NewLoggingDef("github.com/dfuse-io/dlauncher/dashboard.*", nil),
		RegisterFlags: func(cmd *cobra.Command) error {
			cmd.Flags().String("dashboard-grpc-listen-addr", DashboardGrpcServingAddr, "TCP Listener addr for http")
			cmd.Flags().String("dashboard-http-listen-addr", DashboardHTTPListenAddr, "TCP Listener addr for gRPC")
			cmd.Flags().String("dashboard-eos-node-manager-api-addr", EosManagerAPIAddr, "Address of the superviser manager api")
			return nil
		},
		FactoryFunc: func(modules *launcher.RuntimeModules) (launcher.App, error) {
			return dashboard.New(&dashboard.Config{
				HTTPListenAddr:      viper.GetString("dashboard-http-listen-addr"),
				GRPCListenAddr:      viper.GetString("dashboard-grpc-listen-addr"),
				NodeManagerAPIAddr:  viper.GetString("dashboard-eos-node-manager-api-addr"),
				DmeshServiceVersion: viper.GetString("search-common-mesh-service-version"),
				Title:               "dfuse for EOSIO - dashboard",
				BlockExplorerName:   "eosq",
				HeadBlockNumberApp:  "mindreader-node",
			}, &dashboard.Modules{
				Launcher:    modules.Launcher,
				DmeshClient: modules.SearchDmeshClient,
			}), nil
		},
	})
}
