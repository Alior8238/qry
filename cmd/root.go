package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/justestif/qry/internal/config"
	"github.com/justestif/qry/internal/router"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

func Execute(version string) {
	rootCmd := &cobra.Command{
		Use:     "qry <query>",
		Short:   "A terminal-native, agent-first web search hub",
		Version: version,
		Long: `qry routes search queries through pluggable adapter binaries.
Adapters are external executables registered in ~/.config/qry/config.toml.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]

			cfg := &config.Config{}
			if err := viper.Unmarshal(cfg); err != nil {
				return fmt.Errorf("failed to parse config: %w", err)
			}

			if v := viper.GetString("mode"); v != "" {
				cfg.Routing.Mode = v
			}
			if v := viper.GetString("adapter"); v != "" {
				cfg.Routing.Pool = []string{v}
				cfg.Routing.Fallback = nil
			}
			if v := viper.GetInt("num"); v != 0 {
				cfg.Defaults.Num = v
			}

			if cfg.Routing.Mode == "" {
				cfg.Routing.Mode = "first"
			}
			if cfg.Defaults.Num == 0 {
				cfg.Defaults.Num = 10
			}
			if cfg.Defaults.Timeout == 0 {
				cfg.Defaults.Timeout = 5e9 // 5s in nanoseconds
			}

			if err := cfg.Validate(); err != nil {
				return fmt.Errorf("invalid config: %w", err)
			}

			r := router.New(cfg, query)
			output, err := r.Run(context.Background())
			if err != nil {
				if failed, ok := err.(*router.AllAdaptersFailedError); ok {
					failJSON, _ := json.Marshal(failed.FailureOutput())
					fmt.Fprintln(os.Stderr, string(failJSON))
					os.Exit(1)
				}
				return err
			}

			outJSON, err := json.MarshalIndent(output, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to encode output: %w", err)
			}
			fmt.Println(string(outJSON))
			return nil
		},
	}

	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ~/.config/qry/config.toml)")
	rootCmd.Flags().String("adapter", "", "use a specific adapter, bypassing routing")
	rootCmd.Flags().String("mode", "", "routing mode: first or merge (overrides config)")
	rootCmd.Flags().Int("num", 0, "number of results to return (overrides config)")
	rootCmd.Flags().String("timeout", "", "per-adapter timeout e.g. 5s (overrides config)")

	viper.BindPFlag("adapter", rootCmd.Flags().Lookup("adapter"))
	viper.BindPFlag("mode", rootCmd.Flags().Lookup("mode"))
	viper.BindPFlag("num", rootCmd.Flags().Lookup("num"))
	viper.BindPFlag("timeout", rootCmd.Flags().Lookup("timeout"))

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		viper.AddConfigPath(home + "/.config/qry")
		viper.SetConfigName("config")
		viper.SetConfigType("toml")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Fprintln(os.Stderr, "error reading config:", err)
			os.Exit(1)
		}
	}
}
