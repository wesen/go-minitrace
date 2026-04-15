package exporthtml

import (
	"bytes"
	"encoding/json"
	"html/template"
	"strings"
)

type RenderOptions struct {
	PageTitle string
}

type templateData struct {
	PageTitle   string
	EmbeddedCSS template.CSS
	EmbeddedJS  template.JS
	PayloadJSON template.JS
}

func RenderHTML(payload *ReaderExport, opts RenderOptions) ([]byte, error) {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	pageTitle := strings.TrimSpace(opts.PageTitle)
	if pageTitle == "" {
		pageTitle = payload.Session.Title
	}
	if pageTitle == "" {
		pageTitle = payload.Session.ID
	}
	css, err := templateFS.ReadFile("templates/reader.css")
	if err != nil {
		return nil, err
	}
	js, err := templateFS.ReadFile("templates/reader.js")
	if err != nil {
		return nil, err
	}
	tmpl, err := template.ParseFS(templateFS, "templates/reader.html.tmpl")
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, templateData{
		PageTitle:   pageTitle,
		EmbeddedCSS: template.CSS(css),
		EmbeddedJS:  template.JS(js),
		PayloadJSON: template.JS(payloadJSON),
	})
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
