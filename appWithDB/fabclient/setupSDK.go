package fabclient

import (
	reqContext "context"
	"os"
	"strings"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	cliEvent "github.com/hyperledger/fabric-sdk-go/pkg/client/event"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	mspclient "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	contextAPI "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/lookup"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite"
	topFab "github.com/hyperledger/fabric-sdk-go/pkg/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/comm"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	msppkg "github.com/hyperledger/fabric-sdk-go/pkg/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/pathvar"
	"github.com/pkg/errors"
)

type MySetupImpl struct {
	ConfigFile string
	ChannelID  string
	// ChannelConfigFile string
	PeerAdminUser    string
	PeerOrgID        string
	OrdererAdminUser string
	OrdererOrgID     string

	SingleChannelContext contextAPI.ChannelProvider
	WholeChannelContext  contextAPI.ChannelProvider
	SDK                  *fabsdk.FabricSDK
	Identity             msp.Identity
	Targets              []string
	AllTargets           map[string][]string

	AsyncCall bool
	Threshold float64
}

func (setup *MySetupImpl) CloseSDK() {
	if setup.SDK != nil {
		setup.SDK.Close()
		setup.SDK = nil
		setup.SingleChannelContext = nil
		setup.WholeChannelContext = nil
		setup.Identity = nil
		setup.Targets = nil
		setup.AllTargets = nil
	}
}

func (setup *MySetupImpl) CreateOneSDK() (sdk *fabsdk.FabricSDK, err error) {

	configBackend := config.FromFile(pathvar.Subst(setup.ConfigFile))
	one, err := configBackend()
	if err != nil {
		return
	}
	for _, o := range one {
		v, ok := o.Lookup("client.Organization")
		if ok {
			s, ok := v.(string)
			if ok {
				setup.PeerOrgID = s
				break
			}
		}
	}

	setup.CloseSDK()

	sdk, err = fabsdk.New(configBackend)
	if err != nil {
		return
	}

	err = CleanupUserData(sdk)
	if err != nil {
		return
	}

	err = setup.Initialize(sdk)
	if err != nil {
		return
	}

	setup.SingleChannelContext = sdk.ChannelContext(setup.ChannelID,
		fabsdk.WithUser(setup.PeerAdminUser),
		fabsdk.WithOrg(setup.PeerOrgID))
	setup.WholeChannelContext = sdk.ChannelContext(setup.ChannelID,
		fabsdk.WithUser(setup.PeerAdminUser))

	setup.SDK = sdk

	return
}

func (setup *MySetupImpl) NewChannelClient() (*channel.Client, error) {
	return channel.New(setup.SingleChannelContext)
}

func (setup *MySetupImpl) NewSingleLedgerClient() (*ledger.Client, error) {
	return ledger.New(setup.SingleChannelContext)
}

func (setup *MySetupImpl) NewEventClient(opts ...cliEvent.ClientOption) (*cliEvent.Client, error) {
	return cliEvent.New(setup.SingleChannelContext, opts...)
}

func (setup *MySetupImpl) NewWholeLedgerClient() (*ledger.Client, error) {
	return ledger.New(setup.WholeChannelContext)
}

func (setup *MySetupImpl) NewReqContext() (reqContext.Context, reqContext.CancelFunc, error) {
	ctx := setup.SDK.Context(fabsdk.WithUser(setup.PeerAdminUser), fabsdk.WithOrg(setup.PeerOrgID))

	clientContext, err := ctx()
	if err != nil {
		return nil, nil, errors.WithMessage(err, "create context failed")
	}

	reqCtx, cancel := context.NewRequest(&context.Client{Providers: clientContext, SigningIdentity: clientContext}, context.WithTimeoutType(fab.PeerResponse))
	return reqCtx, cancel, nil
}

func (setup *MySetupImpl) GetProposalProcessors(targets []string) ([]fab.ProposalProcessor, error) {
	ctx, err := setup.SingleChannelContext()
	if err != nil {
		return nil, errors.WithMessage(err, "context creation failed")
	}

	var peers []fab.ProposalProcessor
	for _, url := range targets {
		p, err := getPeer(ctx, url)
		if err != nil {
			return nil, err
		}
		peers = append(peers, p)
	}

	return peers, nil
}

// Initialize reads configuration from file and sets up client, channel and event hub
func (setup *MySetupImpl) Initialize(sdk *fabsdk.FabricSDK) error {

	mspClient, err := mspclient.New(sdk.Context(), mspclient.WithOrg(setup.PeerOrgID))
	adminIdentity, err := mspClient.GetSigningIdentity(setup.PeerAdminUser)
	if err != nil {
		return errors.WithMessage(err, "failed to get client context")
	}
	setup.Identity = adminIdentity

	config, err := sdk.Config()
	if err != nil {
		return errors.WithMessage(err, "failed to get client config")
	}

	allTargets, err := allTargetPeers(config, "")
	if err != nil {
		return errors.Wrapf(err, "loading all target peers from config failed")
	}
	setup.AllTargets = allTargets

	targets, err := getTargetsByOrg(allTargets, []string{strings.ToLower(setup.PeerOrgID)})
	if err != nil {
		return errors.Wrapf(err, "loading target peers from config failed")
	}
	setup.Targets = targets

	// r, err := os.Open(setup.ChannelConfigFile)
	// if err != nil {
	// 	return errors.Wrapf(err, "opening channel config file failed")
	// }
	// defer func() {
	// 	if err = r.Close(); err != nil {
	// 		fmt.Printf("close error %v\n", err)
	// 	}
	// }()

	// // Create channel for tests
	// req := resmgmt.SaveChannelRequest{ChannelID: setup.ChannelID, ChannelConfig: r, SigningIdentities: []msp.SigningIdentity{adminIdentity}}
	// if err = InitializeChannel(sdk, setup.PeerAdminUser, setup.PeerOrgID, setup.OrdererAdminUser, setup.OrdererOrgID, req, targets); err != nil {
	// 	return errors.WithMessage(err, "failed to initialize channel")
	// }

	return nil
}

// OrgTargetPeers determines peer endpoints for orgs
func OrgTargetPeers(orgs []string, configBackend ...core.ConfigBackend) ([]string, error) {
	networkConfig := make(map[string]topFab.OrganizationConfig)
	err := lookup.New(configBackend...).UnmarshalKey("organizations", &networkConfig)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get organizations from config ")
	}

	var peers []string
	for _, org := range orgs {
		orgConfig, ok := networkConfig[strings.ToLower(org)]
		if !ok {
			continue
		}
		peers = append(peers, orgConfig.Peers...)
	}
	return peers, nil
}

// AllTargetPeers by prefix
func AllTargetPeers(MSPPrefix string) (ret map[string][]string, err error) {
	if len(mainSetup.AllTargets) > 0 {
		return getAllByPrefix(mainSetup.AllTargets, MSPPrefix)
	}
	configBackend, err := mainSetup.SDK.Config()
	if err != nil {
		return
	}

	return allTargetPeers(configBackend, MSPPrefix)
}

func allTargetPeers(configBackend core.ConfigBackend, MSPPrefix string) (ret map[string][]string, err error) {
	MSPPrefix = strings.ToLower(MSPPrefix)

	networkConfig := make(map[string]topFab.OrganizationConfig)
	err = lookup.New(configBackend).UnmarshalKey("organizations", &networkConfig)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get organizations from config ")
	}
	ret = make(map[string][]string, len(networkConfig))
	for _, v := range networkConfig {
		if len(MSPPrefix) == 0 || strings.HasPrefix(strings.ToLower(v.MSPID), MSPPrefix) {
			ret[strings.ToLower(v.MSPID)] = v.Peers
		}
	}
	return
}

func getAllByPrefix(all map[string][]string, MSPPrefix string) (ret map[string][]string, err error) {
	ret = make(map[string][]string, len(all))
	for k, v := range all {
		if len(MSPPrefix) == 0 || strings.HasPrefix(strings.ToLower(k), MSPPrefix) {
			ret[strings.ToLower(k)] = v
		}
	}

	return
}

// InitializeChannel ...
func InitializeChannel(sdk *fabsdk.FabricSDK, peerUser, orgID, orderUser, ordererorg string, req resmgmt.SaveChannelRequest, targets []string) error {

	joinedTargets, err := FilterTargetsJoinedChannel(sdk, peerUser, orgID, req.ChannelID, targets)
	if err != nil {
		return errors.WithMessage(err, "checking for joined targets failed")
	}

	if len(joinedTargets) != len(targets) {
		_, err := CreateChannel(sdk, orderUser, ordererorg, req)
		if err != nil {
			return errors.Wrapf(err, "create channel failed")
		}

		_, err = JoinChannel(sdk, req.ChannelID, peerUser, orgID)
		if err != nil {
			return errors.Wrapf(err, "join channel failed")
		}
	}
	return nil
}

// FilterTargetsJoinedChannel filters targets to those that have joined the named channel.
func FilterTargetsJoinedChannel(sdk *fabsdk.FabricSDK, adminUser, orgID string, channelID string, targets []string) ([]string, error) {
	var joinedTargets []string

	//prepare context
	clientContext := sdk.Context(fabsdk.WithUser(adminUser), fabsdk.WithOrg(orgID))

	rc, err := resmgmt.New(clientContext)
	if err != nil {
		return nil, errors.WithMessage(err, "failed getting admin user session for org")
	}

	for _, target := range targets {
		// Check if primary peer has joined channel
		alreadyJoined, err := HasPeerJoinedChannel(rc, target, channelID)
		if err != nil {
			return nil, errors.WithMessage(err, "failed while checking if primary peer has already joined channel")
		}
		if alreadyJoined {
			joinedTargets = append(joinedTargets, target)
		}
	}
	return joinedTargets, nil
}

// CreateChannel attempts to save the named channel.
func CreateChannel(sdk *fabsdk.FabricSDK, orderUser, ordererOrgName string, req resmgmt.SaveChannelRequest) (bool, error) {

	//prepare context
	clientContext := sdk.Context(fabsdk.WithUser(orderUser), fabsdk.WithOrg(ordererOrgName))

	// Channel management client is responsible for managing channels (create/update)
	resMgmtClient, err := resmgmt.New(clientContext)
	if err != nil {
		return false, errors.WithMessage(err, "Failed to create new channel management client")
	}

	// Create channel (or update if it already exists)
	if _, err = resMgmtClient.SaveChannel(req, resmgmt.WithRetry(retry.DefaultResMgmtOpts)); err != nil {
		return false, err
	}

	return true, nil
}

// JoinChannel attempts to save the named channel.
func JoinChannel(sdk *fabsdk.FabricSDK, name, adminUser, orgID string) (bool, error) {
	//prepare context
	clientContext := sdk.Context(fabsdk.WithUser(adminUser), fabsdk.WithOrg(orgID))

	// Resource management client is responsible for managing resources (joining channels, install/instantiate/upgrade chaincodes)
	resMgmtClient, err := resmgmt.New(clientContext)
	if err != nil {
		return false, errors.WithMessage(err, "Failed to create new resource management client")
	}

	if err = resMgmtClient.JoinChannel(name, resmgmt.WithRetry(retry.DefaultResMgmtOpts)); err != nil {
		return false, nil
	}
	return true, nil
}

// HasPeerJoinedChannel checks whether the peer has already joined the channel.
// It returns true if it has, false otherwise, or an error
func HasPeerJoinedChannel(client *resmgmt.Client, target string, channel string) (bool, error) {
	foundChannel := false
	response, err := client.QueryChannels(resmgmt.WithTargetEndpoints(target), resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		return false, errors.WithMessage(err, "failed to query channel for peer")
	}
	for _, responseChannel := range response.Channels {
		if responseChannel.ChannelId == channel {
			foundChannel = true
		}
	}

	return foundChannel, nil
}

// CleanupUserData removes user data.
func CleanupUserData(sdk *fabsdk.FabricSDK) (err error) {
	configBackend, err := sdk.Config()
	if err != nil {
		return
	}

	cryptoSuiteConfig := cryptosuite.ConfigFromBackend(configBackend)
	identityConfig, err := msppkg.ConfigFromBackend(configBackend)
	if err != nil {
		return
	}

	keyStorePath := cryptoSuiteConfig.KeyStorePath()
	credentialStorePath := identityConfig.CredentialStorePath()
	err = CleanupTestPath(keyStorePath)
	if err != nil {
		return
	}
	CleanupTestPath(credentialStorePath)
	if err != nil {
		return
	}
	return
}

// CleanupTestPath removes the contents of a state store.
func CleanupTestPath(storePath string) (err error) {
	err = os.RemoveAll(storePath)
	if err != nil {
		return errors.WithMessage(err, "Cleaning up directory "+storePath)
	}
	return
}

func getPeer(ctx contextAPI.Client, url string) (fab.Peer, error) {
	peerCfg, err := comm.NetworkPeerConfig(ctx.EndpointConfig(), url)
	if err != nil {
		return nil, err
	}

	peer, err := ctx.InfraProvider().CreatePeerFromConfig(peerCfg)
	if err != nil {
		return nil, errors.WithMessage(err, "creating peer from config failed")
	}

	return peer, nil
}

func getTargetsByOrg(all map[string][]string, orgname []string) (targets []string, err error) {
	if len(orgname) == 0 {
		return
	}
	targets = make([]string, 0, len(orgname))

Loop:
	for k, v := range all {
		for _, one := range orgname {
			if strings.HasPrefix(strings.ToLower(k), one) {
				if len(v) > 0 {
					targets = append(targets, v...)
				}
				continue Loop
			}
		}
	}

	return
}
