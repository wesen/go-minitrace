package exporthtml

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/go-minitrace/pkg/minitrace"
	queryengine "github.com/go-go-golems/go-minitrace/pkg/query"
	"github.com/pkg/errors"
)

func BuildSessionIndex(archiveGlobs []string) (map[string]string, error) {
	files, err := queryengine.ExpandArchiveGlobs(archiveGlobs)
	if err != nil {
		return nil, err
	}
	index := make(map[string]string, len(files))
	for _, filePath := range files {
		sessionID, err := readSessionIDFromArchive(filePath)
		if err != nil {
			return nil, err
		}
		if previous, ok := index[sessionID]; ok {
			return nil, errors.Errorf("duplicate session ID %q found in %s and %s", sessionID, previous, filePath)
		}
		index[sessionID] = filePath
	}
	if len(index) == 0 {
		return nil, errors.Errorf("archive globs matched no .minitrace.json files")
	}
	return index, nil
}

func LoadSessionByID(index map[string]string, sessionID string) (minitrace.Session, error) {
	sessionPath, ok := index[sessionID]
	if !ok {
		return minitrace.Session{}, errors.Errorf("session not found: %s", sessionID)
	}
	payload, err := os.ReadFile(sessionPath)
	if err != nil {
		return minitrace.Session{}, errors.Wrap(err, "reading session file")
	}
	session := minitrace.Session{}
	if err := json.Unmarshal(payload, &session); err != nil {
		return minitrace.Session{}, errors.Wrap(err, "unmarshaling session file")
	}
	return session, nil
}

func readSessionIDFromArchive(filePath string) (string, error) {
	payload, err := os.ReadFile(filePath)
	if err != nil {
		return "", errors.Wrap(err, "reading session archive for indexing")
	}
	var meta struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(payload, &meta); err != nil {
		return "", errors.Wrap(err, "unmarshaling session archive for indexing")
	}
	if strings.TrimSpace(meta.ID) == "" {
		return "", errors.Errorf("session archive %q is missing id", filepath.Base(filePath))
	}
	return meta.ID, nil
}
