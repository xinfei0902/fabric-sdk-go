package dcmd

import (
	"fmt"

	"../dconfig"
	"../dlog"
	"../tools"
	"github.com/spf13/cobra"
)

// Key for flags
const (
	KeyFlagsDebug    = "debug"
	KeyFlagsHelp     = "help"
	KeyFlagsConfig   = "config"
	KeyFlagsLog      = "log"
	KeyFlagsLogLevel = "loglevel"
)

func isDebug() bool {
	v, ok := dconfig.Get(KeyFlagsDebug)
	if !ok {
		return false
	}
	return dconfig.GetBool(v)
}

func isHelp() bool {
	v, ok := dconfig.Get(KeyFlagsHelp)
	if !ok {
		return false
	}
	return dconfig.GetBool(v)
}

func startLog() error {
	var logfilePath, logLevel string
	logLevel = "warning"
	v, _ := dconfig.Get(KeyFlagsLog)
	logfilePath = dconfig.GetString(v, dlog.GlobalConsoleMark)

	v, _ = dconfig.Get(KeyFlagsLogLevel)
	logLevel = dconfig.GetString(v, "warning")

	fmt.Println("start log: ", logLevel)

	return dlog.InitLog(logfilePath, logLevel, logfilePath != dlog.GlobalConsoleMark)

}

func addBaseFlag() {
	dconfig.SetFileNameKey("c", KeyFlagsConfig, tools.STDPath("config.json"))
	dconfig.Register("h", KeyFlagsHelp, false, "Show this Help")
	dconfig.Register("d", KeyFlagsDebug, false, "Open Debug Model")
	dconfig.Register("l", KeyFlagsLog, dlog.GlobalConsoleMark, "Absolute path of log file")
	dconfig.Register("e", KeyFlagsLogLevel, "warning", "Output Log level: [debug, info, warning, error, fatal, panic]")

	dconfig.Register("", "link", "", "Database Link string")
}

func decorateRunE(cb func(cmd *cobra.Command, args []string) (err error)) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		err := dconfig.ReadFileAndHoldCommand(cmd)
		if err != nil {
			fmt.Println("Quit error: ", err)
			return
		}
		if isHelp() {
			cmd.Help()
			return
		}

		err = startLog()
		if err != nil {
			fmt.Println("Log error: ", err)
			return
		}

		err = cb(cmd, args)
		if err != nil {
			fmt.Println("Quit error: ", err)
		}
	}
}

func makeCommand(name, short, long string,
	configCB func() error,
	cb func(cmd *cobra.Command, args []string) (err error)) *cobra.Command {

	ret := &cobra.Command{
		Use:   name,
		Short: short,
		Long:  long,

		Run: decorateRunE(cb),
	}

	err := configCB()
	if err != nil {
		fmt.Println(err)
		return nil
	}

	dconfig.AddFlags(ret)
	return ret
}
