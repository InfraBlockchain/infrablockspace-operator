// Package chain is package for common chain related functions and types
package chain

// Port is a struct for chain port information
type Port struct {
	RPCPort int32 `json:"rpcPort,omitempty"`
	WSPort  int32 `json:"wsPort,omitempty"`
	P2PPort int32 `json:"p2pPort,omitempty"`
}

// Key is a struct for chain key information
type Key struct {
	KeyType string `json:"type,omitempty"`
	Scheme  string `json:"scheme,omitempty"`
	Seed    string `json:"seed,omitempty"`
}

// KeyStore is volume name for chain key store
const KeyStore string = "chain-keystore"

// Spec is volume name for chain spec
const Spec string = "chain-spec"

// SpecMountPath is volume mount path for chain spec
const SpecMountPath string = "/tmp"

// KeyStoreMountPath is volume mount path for chain key store
const KeyStoreMountPath string = "/keystore"

const DownloadChainSpecImage = "curlimages/curl:8.1.2"

const RelayChainSpecFileName = "relay-chain-spec.json"

const DownloadRelayChainSpecContainer = "download-relay-chain-spec"

const InjectKeysContainer = "inject-keys"

type ChainType string

const RelayChain ChainType = "relay"
const ParaChain ChainType = "para"

const InjectKeyScript string = `
          - |
            set -eu
		{{range .}}
            if [ ! -f/var/run/secrets/{{.KeyType}}/type ]; then
               echo "Error: File/var/run/secrets/{{.KeyType}}/type does not exist"
               exit 1
            fi
           /infra-relay-chain/infrablockspace key insert \
           --keystore-path /keystore \
           --key-type $(cat /var/run/secrets/{{.KeyType}}/type) \
           --scheme $(cat /var/run/secrets/{{.KeyType}}/scheme) \
           --suri /var/run/secrets/{{.KeyType}}/seed \
           && echo "Inserted key into Keystore" \
           || echo "Failed to insert key into Keystore."
		{{end}}
`

const DefaultChainWSPort int32 = 9933
const DefaultChainRPCPort int32 = 9944
const DefaultChainP2PPort int32 = 30333

const VolumeSize100Gi = "100Gi"
const DefaultSecondaryChainWSPort int32 = 9934
const DefaultSecondaryChainRPCPort int32 = 9945
const DefaultSecondaryChainP2PPort int32 = 30334

const SuffixService = "service"
const SuffixPvc = "pvc"
const SuffixHeadlessService = "headless-service"
