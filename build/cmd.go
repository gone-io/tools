package build

import (
	"github.com/gone-io/gonectr/run"
	"github.com/spf13/cobra"
	"os"
)

var Command = &cobra.Command{
	Use:   "build",
	Short: "build gone project",
	Long: "This command will call `gonectr generate ...` to generate gone helper code first, and call `go build` to build gone project." +
		"You can run `go help build` for looking up arguments.",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		return run.GenerateAndRunGoSubCommand("build", os.Args[2:])
	},
}

func init() {
	Command.FParseErrWhitelist.UnknownFlags = true
}
