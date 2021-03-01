package fabclient

import (
	"strings"

	"github.com/pkg/errors"
)

var (
	ErrorUnknownKey         = errors.New("unknown key")
	ErrorSystemResourcePeer = errors.New("Remote peer connect failed")
	ErrorNoTargetPeers      = errors.New("Peers list is empty or wrong")
)

var mainSetup MySetupImpl

func InitGlobalSetup(yamlfile, channelname, orgname, UserName string, async bool, threshold float64) (err error) {
	CloseSetup()

	if len(UserName) == 0 {
		UserName = "Admin"
	}

	if threshold == 0 {
		threshold = 0.4999
	}

	mainSetup = MySetupImpl{
		ChannelID:        channelname,
		PeerAdminUser:    UserName,
		PeerOrgID:        orgname,
		OrdererAdminUser: UserName,
		OrdererOrgID:     "ordererorg",
		ConfigFile:       yamlfile,
		// ChannelConfigFile: channelConfig,
		AsyncCall: async,
		Threshold: threshold,
	}

	_, err = mainSetup.CreateOneSDK()

	return
}

func CloseSetup() {
	mainSetup.CloseSDK()
}

func getPeersFromConfig() (targets []string) {
	return mainSetup.Targets
}

func getAllPeersFromConfig() (allTargets map[string][]string) {
	return mainSetup.AllTargets
}

func stdstring(input string) string {
	return strings.ToLower(strings.TrimSpace(input))
}
