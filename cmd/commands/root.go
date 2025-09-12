package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "tinyapp",
	Short: "TinyApp Service CLI",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.HelpFunc()(cmd, args)
	},
}

func Execute() {
	rootCmd.AddCommand(NewControllerCommand())
	rootCmd.AddCommand(NewGatewayCommand())
	rootCmd.AddCommand(NewServiceCommand())
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
