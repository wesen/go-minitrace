package exporttimeline

import (
	"bytes"
	"html/template"
	"strings"
)

type RenderOptions struct {
	PageTitle string
}

type timelineTemplateData struct {
	PageTitle   string
	EmbeddedCSS template.CSS
	EmbeddedJS  template.JS
	PayloadJSON template.JS
}

func RenderHTML(payload *TimelineExport, opts RenderOptions) ([]byte, error) {
	payloadJSON, err := marshalExportForScriptTag(payload)
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
	if pageTitle == "" {
		pageTitle = "Timeline export"
	}

	css, err := templateFS.ReadFile("templates/timeline.css")
	if err != nil {
		return nil, err
	}
	js, err := templateFS.ReadFile("templates/timeline.js")
	if err != nil {
		return nil, err
	}
	tmpl, err := template.ParseFS(templateFS, "templates/timeline.html.tmpl")
	if err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, timelineTemplateData{
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
