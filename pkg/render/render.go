package render

import (
	"bytes"
	"html/template"

	"github.com/InfraBlockchain/infrablockspace-operator/pkg/chain"
	"github.com/tae2089/bob-logging/logger"
)

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

func RenderingInjectKeyTemplate(Keys []chain.Key) string {
	tmpl, err := template.New("Create Job").Parse(InjectKeyScript)
	if err != nil {
		logger.Error(err)
		return ""
	}
	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, Keys); err != nil {
		logger.Error(err)
		return ""
	}
	return tpl.String()
}
