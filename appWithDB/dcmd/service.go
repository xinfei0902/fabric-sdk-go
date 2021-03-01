package dcmd

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"../dconfig"
	"../dloop"
	"../dservice"
)

func initWebCommand() *cobra.Command {
	return makeCommand(
		"service",
		"Start service",
		`Start a web service for support BaaS actions & request`,
		serviceFlags, serviceRunE)
}

func serviceFlags() (err error) {
	err = addSDKFlags()
	if err != nil {
		return err
	}

	err = addWebFlag()
	if err != nil {
		return err
	}

	err = addCrossFlags()
	if err != nil {
		return err
	}

	err = addWalletFlags()
	if err != nil {
		return err
	}

	err = addMailFlags()
	if err != nil {
		return err
	}
	return nil
}

func serviceRunE(cmd *cobra.Command, args []string) (err error) {
	debug := isDebug()
	link := dconfig.GetStringByKey("link")
	db := len(link) > 0

	// DB
	if true == db {
		err = StartDB(link)
		if err != nil {
			return
		}
	}

	err = startMail()
	if err != nil {
		err = errors.WithMessage(err, "Start mail failed")
		return
	}

	// cross chain
	err = startCrossSDK()
	if err != nil {
		err = errors.WithMessage(err, "Start cross sdk failed")
		return
	}

	// wallet
	err = startWallet()
	if err != nil {
		err = errors.WithMessage(err, "Start wallet failed")
		return
	}

	// sdk
	err = startFabricSDK()
	if err != nil {
		err = errors.WithMessage(err, "Start main sdk failed")
		return
	}

	// time service
	err = dservice.RegisterBackend(debug, db)
	if err != nil {
		err = errors.WithMessage(err, "Register backends failed")
		return
	}
	// web api
	err = dservice.RegisterWebAPI(debug, db)
	if err != nil {
		err = errors.WithMessage(err, "Register WebAPI failed")
		return
	}

	// chaincode api
	err = dservice.RegisterCCAPI(debug, db)
	if err != nil {
		err = errors.WithMessage(err, "Register WebAPI for chaincode failed")
		return
	}

	// web service
	err = dloop.Start()
	if err != nil {
		err = errors.WithMessage(err, "Start backend failed")
		return
	}

	err = StartWeb()

	dloop.Quit()
	dloop.Wait()
	dloop.Clear()

	if err != nil {
		err = errors.WithMessage(err, "Web Quit")
	}
	return err
}
