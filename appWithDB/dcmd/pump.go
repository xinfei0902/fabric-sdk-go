package dcmd

import (
	"../dconfig"
	"../dloop"
	"../dservice"
	"github.com/spf13/cobra"
)

func initPumpCommand() *cobra.Command {
	return makeCommand(
		"pump",
		"Start pump service",
		`Start service pump infomation from chain into DB`,
		pumpFlags, pumpRunE)
}

func pumpFlags() (err error) {
	err = addSDKFlags()
	if err != nil {
		return err
	}

	err = addWebFlag()
	if err != nil {
		return err
	}
	return nil
}

func pumpRunE(cmd *cobra.Command, args []string) (err error) {
	debug := isDebug()
	link := dconfig.GetStringByKey("link")

	// DB
	if len(link) > 0 {
		err = StartDB(link)
		if err != nil {
			return
		}
	}

	// sdk
	err = startFabricSDK()
	if err != nil {
		return
	}

	// time service
	err = dservice.RegisterBackend(debug, len(link) > 0)
	if err != nil {
		return
	}

	// web api
	err = dservice.RegisterWebAPI(debug, len(link) > 0)
	if err != nil {
		return
	}

	// web service
	err = dloop.Start()
	if err != nil {
		return
	}

	err = StartWeb()

	dloop.Quit()
	dloop.Wait()
	dloop.Clear()

	return err
}
