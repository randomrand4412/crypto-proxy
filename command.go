package cryptoproxy

import (
	"fmt"

	"github.com/spf13/cobra"
)

func CreateCommand() *cobra.Command {
	config := &Config{}

	command := &cobra.Command{
		Use:     `proxy`,
		Short:   `reliable proxy`,
		Long:    `No, really, very reliable proxy`,
		PreRunE: FlagsValidator(config),
		RunE: func(cmd *cobra.Command, args []string) error {
			return Start(config)
		},
	}

	command.PersistentFlags().IntVarP(&config.Port, "port", "p", 8080, "server port")
	command.PersistentFlags().StringVarP(&config.AuthToken, "token", "t", "", "auth token")
	command.PersistentFlags().StringVarP(&config.OriginServer, "origin", "o", "", "origin server url")
	command.PersistentFlags().IntVar(&config.RequestTimeoutMs, "ingress-request-timeout-ms", 2000, "time to keep ingress connection alive")
	command.PersistentFlags().StringSliceVar(&config.Paths, "paths", []string{}, "list of origin paths to proxy")

	return command
}

func FlagsValidator(cfg *Config) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if len(cfg.OriginServer) == 0 {
			return fmt.Errorf("origin cannot be empty")
		}
		if len(cfg.Paths) == 0 {
			return fmt.Errorf("paths cannot be empty")
		}
		if len(cfg.AuthToken) == 0 {
			return fmt.Errorf("token cannot be empty")
		}

		return nil
	}
}
