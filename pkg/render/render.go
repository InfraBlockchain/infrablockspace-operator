package render

import (
	"bytes"
	"github.com/tae2089/bob-logging/logger"
	"html/template"
)

func RenderingInTemplate(InjectKeyScript string, data any) string {
	tmpl, err := template.New("Create Job").Parse(InjectKeyScript)
	if err != nil {
		logger.Error(err)
		return ""
	}
	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, data); err != nil {
		logger.Error(err)
		return ""
	}
	return tpl.String()
}
