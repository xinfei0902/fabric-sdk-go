package fabclient

import (
	cb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	mb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/msp"
)

type ConfigEnvelope struct {
	Config     *Config   `json:"config,omitempty"`
	LastUpdate *Envelope `json:"last_update,omitempty"`
}

// ConfigValue represents an individual piece of config data
type ConfigValue struct {
	Version   uint64      `json:"version,omitempty"`
	Value     interface{} `json:"value,omitempty"`
	ModPolicy string      `json:"mod_policy,omitempty"`
}

// Config represents the config for a particular channel
type Config struct {
	Sequence     uint64       `json:"sequence,omitempty"`
	ChannelGroup *ConfigGroup `json:"channel_group,omitempty"`
}

// ConfigGroup is the hierarchical data structure for holding config
type ConfigGroup struct {
	Version   uint64                   `json:"version,omitempty"`
	Groups    map[string]*ConfigGroup  `json:"groups,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	Values    map[string]*ConfigValue  `json:"values,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	Policies  map[string]*ConfigPolicy `json:"policies,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	ModPolicy string                   `json:"mod_policy,omitempty"`
}

type ConfigPolicy struct {
	Version   uint64  `json:"version,omitempty"`
	Policy    *Policy `json:"policy,omitempty"`
	ModPolicy string  `json:"mod_policy,omitempty"`
}

// Policy expresses a policy which the orderer can evaluate, because there has been some desire expressed to support
// multiple policy engines, this is typed as a oneof for now
type Policy struct {
	Type  int32       `json:"type,omitempty"`
	Value interface{} `json:"value,omitempty"`
}

// Envelope wraps a Payload with a signature so that the message may be authenticated
type Envelope struct {
	// A marshaled Payload
	Payload interface{} `json:"payload,omitempty"`
	// A signature by the creator specified in the Payload header
	Signature []byte `json:"signature,omitempty"`
}

type MSPConfig struct {
	// Type holds the type of the MSP; the default one would
	// be of type FABRIC implementing an X.509 based provider
	Type int32 `json:"type,omitempty"`
	// Config is MSP dependent configuration info
	Config interface{} `json:"config,omitempty"`
}

type FabricMSPConfig struct {
	// Name holds the identifier of the MSP; MSP identifier
	// is chosen by the application that governs this MSP.
	// For example, and assuming the default implementation of MSP,
	// that is X.509-based and considers a single Issuer,
	// this can refer to the Subject OU field or the Issuer OU field.
	Name string `json:"name,omitempty"`
	// List of root certificates trusted by this MSP
	// they are used upon certificate validation (see
	// comment for IntermediateCerts below)
	RootCerts []string `json:"root_certs,omitempty"`
	// List of intermediate certificates trusted by this MSP;
	// they are used upon certificate validation as follows:
	// validation attempts to build a path from the certificate
	// to be validated (which is at one end of the path) and
	// one of the certs in the RootCerts field (which is at
	// the other end of the path). If the path is longer than
	// 2, certificates in the middle are searched within the
	// IntermediateCerts pool
	IntermediateCerts []string `json:"intermediate_certs,omitempty"`
	// Identity denoting the administrator of this MSP
	Admins []string `json:"admins,omitempty"`
	// Identity revocation list
	RevocationList []string `json:"revocation_list,omitempty"`
	// SigningIdentity holds information on the signing identity
	// this peer is to use, and which is to be imported by the
	// MSP defined before
	SigningIdentity *SigningIdentityInfo `json:"signing_identity,omitempty"`
	// OrganizationalUnitIdentifiers holds one or more
	// fabric organizational unit identifiers that belong to
	// this MSP configuration
	OrganizationalUnitIdentifiers []*FabricOUIdentifier `json:"organizational_unit_identifiers,omitempty"`
	// FabricCryptoConfig contains the configuration parameters
	// for the cryptographic algorithms used by this MSP
	CryptoConfig *FabricCryptoConfig `json:"crypto_config,omitempty"`
	// List of TLS root certificates trusted by this MSP.
	// They are returned by GetTLSRootCerts.
	TlsRootCerts []string `json:"tls_root_certs,omitempty"`
	// List of TLS intermediate certificates trusted by this MSP;
	// They are returned by GetTLSIntermediateCerts.
	TlsIntermediateCerts []string `json:"tls_intermediate_certs,omitempty"`
	// fabric_node_ous contains the configuration to distinguish clients from peers from orderers
	// based on the OUs.
	FabricNodeOus *FabricNodeOUs `json:"fabric_node_ous,omitempty"`
}

type SigningIdentityInfo struct {
	// PublicSigner carries the public information of the signing
	// identity. For an X.509 provider this would be represented by
	// an X.509 certificate
	PublicSigner string `json:"public_signer,omitempty"`
	// PrivateSigner denotes a reference to the private key of the
	// peer's signing identity
	PrivateSigner *KeyInfo `json:"private_signer,omitempty"`
}

// FabricOUIdentifier represents an organizational unit and
// its related chain of trust identifier.
type FabricOUIdentifier struct {
	// Certificate represents the second certificate in a certification chain.
	// (Notice that the first certificate in a certification chain is supposed
	// to be the certificate of an identity).
	// It must correspond to the certificate of root or intermediate CA
	// recognized by the MSP this message belongs to.
	// Starting from this certificate, a certification chain is computed
	// and bound to the OrganizationUnitIdentifier specified
	Certificate string `json:"certificate,omitempty"`
	// OrganizationUnitIdentifier defines the organizational unit under the
	// MSP identified with MSPIdentifier
	OrganizationalUnitIdentifier string `json:"organizational_unit_identifier,omitempty"`
}

// KeyInfo represents a (secret) key that is either already stored
// in the bccsp/keystore or key material to be imported to the
// bccsp key-store. In later versions it may contain also a
// keystore identifier
type KeyInfo struct {
	// Identifier of the key inside the default keystore; this for
	// the case of Software BCCSP as well as the HSM BCCSP would be
	// the SKI of the key
	KeyIdentifier string `json:"key_identifier,omitempty"`
	// KeyMaterial (optional) for the key to be imported; this is
	// properly encoded key bytes, prefixed by the type of the key
	KeyMaterial string `json:"key_material,omitempty"`
}

// FabricCryptoConfig contains configuration parameters
// for the cryptographic algorithms used by the MSP
// this configuration refers to
type FabricCryptoConfig struct {
	// SignatureHashFamily is a string representing the hash family to be used
	// during sign and verify operations.
	// Allowed values are "SHA2" and "SHA3".
	SignatureHashFamily string `json:"signature_hash_family,omitempty"`
	// IdentityIdentifierHashFunction is a string representing the hash function
	// to be used during the computation of the identity identifier of an MSP identity.
	// Allowed values are "SHA256", "SHA384" and "SHA3_256", "SHA3_384".
	IdentityIdentifierHashFunction string `json:"identity_identifier_hash_function,omitempty"`
}

// FabricNodeOUs contains configuration to tell apart clients from peers from orderers
// based on OUs. If NodeOUs recognition is enabled then an msp identity
// that does not contain any of the specified OU will be considered invalid.
type FabricNodeOUs struct {
	// If true then an msp identity that does not contain any of the specified OU will be considered invalid.
	Enable bool `json:"enable,omitempty"`
	// OU Identifier of the clients
	ClientOuIdentifier *FabricOUIdentifier `json:"client_ou_identifier,omitempty"`
	// OU Identifier of the peers
	PeerOuIdentifier *FabricOUIdentifier `json:"peer_ou_identifier,omitempty"`
}

// IdemixMSPConfig collects all the configuration information for
// an Idemix MSP.
type IdemixMSPConfig struct {
	// Name holds the identifier of the MSP
	Name string `json:"name,omitempty"`
	// ipk represents the (serialized) issuer public key
	Ipk string `json:"ipk,omitempty"`
	// signer may contain crypto material to configure a default signer
	Signer *IdemixMSPSignerConfig `json:"signer,omitempty"`
	// revocation_pk is the public key used for revocation of credentials
	RevocationPk string `json:"revocation_pk,omitempty"`
	// epoch represents the current epoch (time interval) used for revocation
	Epoch int64 `json:"epoch,omitempty"`
}

// IdemixMSPSIgnerConfig contains the crypto material to set up an idemix signing identity
type IdemixMSPSignerConfig struct {
	// cred represents the serialized idemix credential of the default signer
	Cred string `json:"cred,omitempty"`
	// sk is the secret key of the default signer, corresponding to credential Cred
	Sk string `json:"sk,omitempty"`
	// organizational_unit_identifier defines the organizational unit the default signer is in
	OrganizationalUnitIdentifier string `json:"organizational_unit_identifier,omitempty"`
	// role defines whether the default signer is admin, peer, member or client
	Role int32 `protobuf:"varint,4,opt,name=role" json:"role,omitempty"`
	// enrollment_id contains the enrollment id of this signer
	EnrollmentId string `json:"enrollment_id,omitempty"`
	// credential_revocation_information contains a serialized CredentialRevocationInformation
	CredentialRevocationInformation string `json:"credential_revocation_information,omitempty"`
}

// SignaturePolicyEnvelope wraps a SignaturePolicy and includes a version for future enhancements
type SignaturePolicyEnvelope struct {
	Version    int32               `json:"version,omitempty"`
	Rule       *cb.SignaturePolicy `json:"rule,omitempty"`
	Identities []*MSPPrincipal     `json:"identities,omitempty"`
}

type MSPPrincipal struct {
	// Classification describes the way that one should process
	// Principal. An Classification value of "ByOrganizationUnit" reflects
	// that "Principal" contains the name of an organization this MSP
	// handles. A Classification value "ByIdentity" means that
	// "Principal" contains a specific identity. Default value
	// denotes that Principal contains one of the groups by
	// default supported by all MSPs ("admin" or "member").
	PrincipalClassification mb.MSPPrincipal_Classification `json:"principal_classification,omitempty"`
	// Principal completes the policy principal definition. For the default
	// principal types, Principal can be either "Admin" or "Member".
	// For the ByOrganizationUnit/ByIdentity values of Classification,
	// PolicyPrincipal acquires its value from an organization unit or
	// identity, respectively.
	// For the Combined Classification type, the Principal is a marshalled
	// CombinedPrincipal.
	Principal interface{} `json:"principal,omitempty"`
}

type CombinedPrincipal struct {
	// Principals refer to combined principals
	Principals []*MSPPrincipal `json:"principals,omitempty"`
}

type ConfigUpdateEnvelope struct {
	ConfigUpdate interface{}           `json:"config_update,omitempty"`
	Signatures   []*cb.ConfigSignature `json:"signatures,omitempty"`
}

type ConfigUpdate struct {
	ChannelId    string             `json:"channel_id,omitempty"`
	ReadSet      *ConfigGroup       `json:"read_set,omitempty"`
	WriteSet     *ConfigGroup       `json:"write_set,omitempty"`
	IsolatedData map[string]*Config `json:"isolated_data,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value,proto3"`
}
