package commands

import (
	"github.com/spf13/cobra"
	server "github.com/tinymultiverse/tinyapp/server/cmd"
)

func NewServiceCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "server",
		Short: "Start the server",
		Run: func(cmd *cobra.Command, args []string) {
			server.Start()
		},
	}
}
