package fabclient

import (
	"github.com/golang/protobuf/proto"
	cb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	mb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/msp"
	ab "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/orderer"
	pp "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"

	"../dlog"
)

func ParseConfigEnvelope(input []byte) (ret *ConfigEnvelope, err error) {
	ret = &ConfigEnvelope{}

	// Todo
	configEnvelope := &cb.ConfigEnvelope{}
	if err := proto.Unmarshal(input, configEnvelope); err != nil {
		return nil, errors.Wrap(err, "unmarshal config envelope failed")
	}
	// 1
	ret.Config, err = ParseConfig(configEnvelope.GetConfig())
	if err != nil {
		return nil, err
	}

	// 2
	ret.LastUpdate, err = ParseConfigLastUpdate(configEnvelope.GetLastUpdate())

	return
}

func ParseConfig(conf *cb.Config) (ret *Config, err error) {
	ret = &Config{}

	ret.Sequence = conf.GetSequence()

	channelGroup := conf.GetChannelGroup()

	ret.ChannelGroup, err = ParseChannelGroup(channelGroup)

	return
}

func ParseChannelGroup(group *cb.ConfigGroup) (ret *ConfigGroup, err error) {
	ret = &ConfigGroup{}
	ret.ModPolicy = group.GetModPolicy()
	ret.Version = group.GetVersion()
	ret.Policies = make(map[string]*ConfigPolicy, len(group.GetPolicies()))
	for k, v := range group.GetPolicies() {
		ret.Policies[k], err = ParsePolicy(v)
		if err != nil {
			return nil, err
		}
	}

	ret.Values = make(map[string]*ConfigValue, len(group.GetValues()))
	for k, v := range group.GetValues() {
		ret.Values[k], err = ParseValues(k, v)
		if err != nil {
			return nil, err
		}
	}

	ret.Groups = make(map[string]*ConfigGroup, len(group.GetGroups()))
	for k, v := range group.GetGroups() {
		ret.Groups[k], err = ParseChannelGroup(v)
		if err != nil {
			return nil, err
		}
	}

	return
}

func ToFabricOUIdentifier(input *mb.FabricOUIdentifier) *FabricOUIdentifier {
	if nil == input {
		return nil
	}
	return &FabricOUIdentifier{
		Certificate:                  string(input.GetCertificate()),
		OrganizationalUnitIdentifier: input.GetOrganizationalUnitIdentifier(),
	}
}

func ParseValues(k string, v *cb.ConfigValue) (ret *ConfigValue, err error) {
	ret = &ConfigValue{}
	ret.Version = v.GetVersion()
	ret.ModPolicy = v.GetModPolicy()

	switch k {
	case BatchSizeKey:
		one := &ab.BatchSize{}

		err = proto.Unmarshal(v.GetValue(), one)
		if err != nil {
			return nil, err
		}
		ret.Value = one
	case AnchorPeersKey:
		one := &pp.AnchorPeers{}

		err = proto.Unmarshal(v.GetValue(), one)
		if err != nil {
			return nil, err
		}
		ret.Value = one

	case ConsensusTypeKey:
		one := &ab.ConsensusType{}
		err = proto.Unmarshal(v.GetValue(), one)
		if err != nil {
			return nil, err
		}
		ret.Value = one
	case BatchTimeoutKey:
		one := &ab.BatchTimeout{}
		err = proto.Unmarshal(v.GetValue(), one)
		if err != nil {
			return nil, err
		}
		ret.Value = one
	case ChannelRestrictionsKey:
		one := &ab.ChannelRestrictions{}
		err = proto.Unmarshal(v.GetValue(), one)
		if err != nil {
			return nil, err
		}
		ret.Value = one
	case HashingAlgorithmKey:
		one := &cb.HashingAlgorithm{}
		err = proto.Unmarshal(v.GetValue(), one)
		if err != nil {
			return nil, err
		}
		ret.Value = one
	case BlockDataHashingStructureKey:
		one := &cb.BlockDataHashingStructure{}
		err = proto.Unmarshal(v.GetValue(), one)
		if err != nil {
			return nil, err
		}
		ret.Value = one
	case CapabilitiesKey:
		one := &cb.Capabilities{}
		err = proto.Unmarshal(v.GetValue(), one)
		if err != nil {
			return nil, err
		}
		ret.Value = one
	case KafkaBrokersKey:
		one := &ab.KafkaBrokers{}
		err = proto.Unmarshal(v.GetValue(), one)
		if err != nil {
			return nil, err
		}
		ret.Value = one
	case ConsortiumKey:
		one := &cb.Consortium{}
		err = proto.Unmarshal(v.GetValue(), one)
		if err != nil {
			return nil, err
		}
		ret.Value = one
	case OrdererAddressesKey:
		one := &cb.OrdererAddresses{}
		err = proto.Unmarshal(v.GetValue(), one)
		if err != nil {
			return nil, err
		}
		ret.Value = one
	case MSPKey:
		one := MSPConfig{}
		tmp := &mb.MSPConfig{}
		err = proto.Unmarshal(v.GetValue(), tmp)
		if err != nil {
			return nil, err
		}
		one.Type = tmp.GetType()

		switch ProviderType(tmp.GetType()) {
		case FABRIC:
			tmp2 := &mb.FabricMSPConfig{}
			two := &FabricMSPConfig{}

			err = proto.Unmarshal(tmp.GetConfig(), tmp2)
			if err != nil {
				return nil, err
			}

			two.Admins = make([]string, len(tmp2.GetAdmins()))
			for i, t := range tmp2.GetAdmins() {
				two.Admins[i] = string(t)
			}

			if nil != tmp2.GetCryptoConfig() {
				two.CryptoConfig = &FabricCryptoConfig{
					SignatureHashFamily:            tmp2.GetCryptoConfig().GetSignatureHashFamily(),
					IdentityIdentifierHashFunction: tmp2.GetCryptoConfig().GetIdentityIdentifierHashFunction(),
				}
			}

			if nil != tmp2.GetFabricNodeOus() {
				p := tmp2.GetFabricNodeOus()
				two.FabricNodeOus = &FabricNodeOUs{
					Enable:             p.GetEnable(),
					ClientOuIdentifier: ToFabricOUIdentifier(p.GetClientOuIdentifier()),
					PeerOuIdentifier:   ToFabricOUIdentifier(p.GetPeerOuIdentifier()),
				}
			}

			two.IntermediateCerts = make([]string, len(tmp2.GetIntermediateCerts()))

			for i, t := range tmp2.GetIntermediateCerts() {
				two.IntermediateCerts[i] = string(t)
			}

			two.Name = tmp2.GetName()
			two.OrganizationalUnitIdentifiers = make([]*FabricOUIdentifier, len(tmp2.GetOrganizationalUnitIdentifiers()))
			for i, t := range tmp2.GetOrganizationalUnitIdentifiers() {
				two.OrganizationalUnitIdentifiers[i] = ToFabricOUIdentifier(t)
			}

			two.RevocationList = make([]string, len(tmp2.GetRevocationList()))
			for i, t := range tmp2.GetRevocationList() {
				two.RevocationList[i] = string(t)
			}

			if nil != tmp2.GetSigningIdentity() {
				p := tmp2.GetSigningIdentity()

				two.SigningIdentity = &SigningIdentityInfo{
					PublicSigner: string(p.GetPublicSigner()),
				}
				if nil != p.GetPrivateSigner() {
					two.SigningIdentity.PrivateSigner = &KeyInfo{
						KeyIdentifier: p.GetPrivateSigner().GetKeyIdentifier(),
						KeyMaterial:   string(p.GetPrivateSigner().GetKeyMaterial()),
					}
				}
			}

			two.TlsIntermediateCerts = make([]string, len(tmp2.GetTlsIntermediateCerts()))

			for i, t := range tmp2.GetTlsIntermediateCerts() {
				two.TlsIntermediateCerts[i] = string(t)
			}

			two.TlsRootCerts = make([]string, len(tmp2.GetTlsRootCerts()))
			for i, t := range tmp2.GetTlsRootCerts() {
				two.TlsRootCerts[i] = string(t)
			}

			one.Config = two
		case IDEMIX:
			tmp2 := &mb.IdemixMSPConfig{}
			two := &IdemixMSPConfig{}
			err = proto.Unmarshal(tmp.GetConfig(), tmp2)
			if err != nil {
				return nil, err
			}

			two.Epoch = tmp2.GetEpoch()
			two.Ipk = string(tmp2.GetIpk())
			two.Name = tmp2.GetName()
			two.RevocationPk = string(tmp2.GetRevocationPk())
			if nil != tmp2.GetSigner() {
				p := tmp2.GetSigner()
				two.Signer = &IdemixMSPSignerConfig{
					Cred: string(p.GetCred()),
					Sk:   string(p.GetSk()),
					OrganizationalUnitIdentifier: p.GetOrganizationalUnitIdentifier(),
					Role:                            p.GetRole(),
					EnrollmentId:                    p.GetEnrollmentId(),
					CredentialRevocationInformation: string(p.GetCredentialRevocationInformation()),
				}
			}

			one.Config = two
		default:
			dlog.Debugf("Unknown ConfigValue MSPconfig Type: %v", tmp.GetType())
		}
		ret.Value = one
	default:
		dlog.Debugf("Unknown ConfigValue key: %v", k)
	}

	return
}

func ToMSPPrincipal(input *mb.MSPPrincipal) (*MSPPrincipal, error) {
	if nil == input {
		return nil, nil
	}
	ret := &MSPPrincipal{}
	ret.PrincipalClassification = input.GetPrincipalClassification()
	buff := input.GetPrincipal()
	if len(buff) == 0 {
		return ret, nil
	}
	switch input.GetPrincipalClassification() {
	case mb.MSPPrincipal_ROLE:
		three := &mb.MSPRole{}
		err := proto.Unmarshal(buff, three)
		if err != nil {
			return nil, err
		}

		ret.Principal = three
	case mb.MSPPrincipal_ORGANIZATION_UNIT:
		three := &mb.OrganizationUnit{}
		err := proto.Unmarshal(buff, three)
		if err != nil {
			return nil, err
		}

		ret.Principal = three
	case mb.MSPPrincipal_IDENTITY:
		three := &mb.SerializedIdentity{}
		err := proto.Unmarshal(buff, three)
		if err != nil {
			return nil, err
		}
		ret.Principal = three
	case mb.MSPPrincipal_ANONYMITY:
		three := &mb.MSPIdentityAnonymity{}
		err := proto.Unmarshal(buff, three)
		if err != nil {
			return nil, err
		}
		ret.Principal = three
	case mb.MSPPrincipal_COMBINED:
		tmp3 := &mb.CombinedPrincipal{}
		err := proto.Unmarshal(buff, tmp3)
		if err != nil {
			return nil, err
		}

		three := &CombinedPrincipal{}
		three.Principals = make([]*MSPPrincipal, len(tmp3.GetPrincipals()))
		for i, t := range tmp3.GetPrincipals() {
			three.Principals[i], err = ToMSPPrincipal(t)
			if err != nil {
				return nil, err
			}
		}
		ret.Principal = three
	default:
		dlog.Debugf("Unknown policy MSPPrincipal key: %v", input.GetPrincipalClassification())
	}
	return ret, nil
}

func ParsePolicy(v *cb.ConfigPolicy) (ret *ConfigPolicy, err error) {
	ret = &ConfigPolicy{}
	ret.ModPolicy = v.GetModPolicy()
	ret.Version = v.GetVersion()

	policy := v.GetPolicy()
	if policy == nil {
		return
	}
	ret.Policy = &Policy{
		Type: policy.GetType(),
	}

	switch cb.Policy_PolicyType(policy.GetType()) {
	case cb.Policy_SIGNATURE:
		// Todo here
		tmp := &cb.SignaturePolicyEnvelope{}

		err := proto.Unmarshal(policy.GetValue(), tmp)
		if err != nil {
			return nil, err
		}

		one := &SignaturePolicyEnvelope{}

		one.Version = tmp.GetVersion()
		one.Rule = tmp.GetRule()

		one.Identities = make([]*MSPPrincipal, len(tmp.GetIdentities()))

		for i, t := range tmp.GetIdentities() {
			one.Identities[i], err = ToMSPPrincipal(t)
			if err != nil {
				return nil, errors.Wrap(err, "unmarshal signature policy from config failed")
			}
		}

		ret.Policy.Value = one
	case cb.Policy_UNKNOWN:
	case cb.Policy_MSP:
	case cb.Policy_IMPLICIT_META:
		one := &cb.ImplicitMetaPolicy{}
		err := proto.Unmarshal(policy.GetValue(), one)
		if err != nil {
			return nil, errors.Wrap(err, "unmarshal implicit meta policy from config failed")
		}

		ret.Policy.Value = one
	default:
		dlog.Debugf("Unknown Policy type: %v", policy.GetType())
	}

	return
}

func ParseConfigLastUpdate(envelope *cb.Envelope) (ret *Envelope, err error) {

	ret = &Envelope{}

	ret.Signature = envelope.GetSignature()

	payload := &cb.Payload{}
	err = proto.Unmarshal(envelope.GetPayload(), payload)
	if err != nil {
		return nil, err
	}

	tmp := &cb.ConfigUpdateEnvelope{}
	one := ConfigUpdateEnvelope{}
	err = proto.Unmarshal(payload.GetData(), tmp)
	if err != nil {
		return nil, err
	}

	one.Signatures = tmp.GetSignatures()

	tmp2 := &cb.ConfigUpdate{}
	two := &ConfigUpdate{}
	err = proto.Unmarshal(tmp.GetConfigUpdate(), tmp2)
	if err != nil {
		return nil, err
	}

	two.ChannelId = tmp2.GetChannelId()
	two.ReadSet, err = ParseChannelGroup(tmp2.GetReadSet())
	if err != nil {
		return nil, err
	}
	two.WriteSet, err = ParseChannelGroup(tmp2.GetWriteSet())
	if err != nil {
		return nil, err
	}

	two.IsolatedData = make(map[string]*Config, len(tmp2.GetIsolatedData()))

	for k, v := range tmp2.GetIsolatedData() {

		tmp3 := &cb.Config{}
		err = proto.Unmarshal(v, tmp3)
		three := &Config{}
		three.Sequence = tmp3.GetSequence()
		three.ChannelGroup, err = ParseChannelGroup(tmp3.GetChannelGroup())

		two.IsolatedData[k] = three

	}

	one.ConfigUpdate = two

	ret.Payload = one

	return
}
