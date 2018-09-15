package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "markovcg <train|generate>",
	Short: "Markov chain text generator",
	Long:  ``,
}

func init() {
	rootCmd.AddCommand(trainCmd)
	rootCmd.AddCommand(formatCmd)
}

// Execute is the main function of Cobra
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
