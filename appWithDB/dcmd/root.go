package dcmd

import (
	"github.com/spf13/cobra"
)

var globalCommand = &cobra.Command{
	Use:   "root",
	Short: "root",
	Long:  "root",
}

func init() {
	addBaseFlag()
	// append service
	globalCommand.AddCommand(initWebCommand())
	globalCommand.AddCommand(initPumpCommand())
}

// Execute tree command parse and actions
func Execute() error {
	return globalCommand.Execute()
}
