package exporthtml

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	reModuleScript       = regexp.MustCompile(`(?s)<script\s+type="module"[^>]*src="([^"]+)"[^>]*></script>`)
	reRemainingScriptSrc = regexp.MustCompile(`(?s)<script[^>]*\ssrc="[^"]+"[^>]*></script>`)
	reLinkHref           = regexp.MustCompile(`(?s)<link\b[^>]*href="[^"]+"[^>]*>`)
	reJSONScript         = regexp.MustCompile(`(?s)<script\s+id="minitrace-export-data"\s+type="application/json">.*?</script>`)
	reTitle              = regexp.MustCompile(`(?s)<title>.*?</title>`)
)

func RenderHTMLFromBuiltBundle(payload *ReaderExport, distDir string, opts RenderOptions) ([]byte, error) {
	payloadJSON, err := marshalPayloadForScriptTag(payload)
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
	htmlPath := filepath.Join(distDir, "export-reader.html")
	htmlBytes, err := os.ReadFile(htmlPath)
	if err != nil {
		return nil, err
	}
	html := string(htmlBytes)
	matches := reModuleScript.FindStringSubmatch(html)
	if len(matches) != 2 {
		return nil, os.ErrInvalid
	}
	jsPath := filepath.Join(distDir, strings.TrimPrefix(matches[1], "/"))
	jsBytes, err := os.ReadFile(jsPath)
	if err != nil {
		return nil, err
	}
	html = reModuleScript.ReplaceAllLiteralString(html, `<script type="module">`+string(jsBytes)+`</script>`)
	html = reRemainingScriptSrc.ReplaceAllString(html, "")
	html = reLinkHref.ReplaceAllString(html, "")
	html = reJSONScript.ReplaceAllLiteralString(html, `<script id="minitrace-export-data" type="application/json">`+payloadJSON+`</script>`)
	html = reTitle.ReplaceAllLiteralString(html, `<title>`+htmlEscapeTitle(pageTitle)+`</title>`)
	return []byte(html), nil
}

func htmlEscapeTitle(s string) string {
	replacer := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
	return replacer.Replace(s)
}
