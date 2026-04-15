package exporthtml

import (
	"encoding/json"
	"strings"
)

var scriptTagJSONReplacer = strings.NewReplacer(
	"<", `\u003c`,
	">", `\u003e`,
	"&", `\u0026`,
	"\u2028", `\u2028`,
	"\u2029", `\u2029`,
)

func marshalPayloadForScriptTag(payload *ReaderExport) (string, error) {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return scriptTagJSONReplacer.Replace(string(payloadJSON)), nil
}
