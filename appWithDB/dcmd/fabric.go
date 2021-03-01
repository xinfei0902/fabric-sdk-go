package dcmd

import (
	"../dconfig"
	"../fabclient"
)

func addSDKFlags() (err error) {
	dconfig.Register("y", "yaml", "core.yaml", "For Network/organization config")
	dconfig.Register("n", "channelname", "mychannel", "For Channel ID")
	dconfig.Register("o", "org", "Org1", "Organization name in network")
	dconfig.Register("u", "user", "Admin", "Organization user name in network")
	dconfig.Register("", "asynccall", false, "Async Call model")

	dconfig.Register("", "threshold", 0.4999, "Threshold: count of peers reach a consensus")

	return nil
}

func startFabricSDK() error {
	YamlConfigPath := dconfig.GetStringByKey("yaml")
	Channel := dconfig.GetStringByKey("channelname")
	Org := dconfig.GetStringByKey("org")
	User := dconfig.GetStringByKey("user")
	async := dconfig.GetBoolByKey("asynccall")

	threshold := dconfig.GetFloatByKey("threshold")

	return fabclient.InitGlobalSetup(YamlConfigPath, Channel, Org, User, async, threshold)
}
