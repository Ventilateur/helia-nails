package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// cookiesCmd represents the cookies command
var cookiesCmd = &cobra.Command{
	Use: "cookies",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("cookies called")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(cookiesCmd)
}
