package main

import "github.com/spf13/cobra"

var rootCmd = cobra.Command{
	Use: "quictool",
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

func main() {
	_ = rootCmd.Execute()
}
