package commands

import (
	"github.com/spf13/cobra"
	gateway "github.com/tinymultiverse/tinyapp/gateway/cmd"
)

func NewGatewayCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "gateway",
		Short: "Start the gateway (proxy)",
		Run: func(cmd *cobra.Command, args []string) {
			gateway.Start()
		},
	}
}
