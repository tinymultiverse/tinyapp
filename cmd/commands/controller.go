package commands

import (
	"github.com/spf13/cobra"
	controller "github.com/tinymultiverse/tinyapp/controller/cmd"
)

func NewControllerCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "controller",
		Short: "Start the controller",
		Run: func(cmd *cobra.Command, args []string) {
			controller.Start()
		},
	}
}
