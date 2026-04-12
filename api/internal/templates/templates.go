package templates

import (
	"embed"
	"html/template"
)

//go:embed email/*.html
var emailTemplates embed.FS

func ParseEmailTemplates() (*template.Template, error) {
	return template.ParseFS(emailTemplates, "email/*.html")
}