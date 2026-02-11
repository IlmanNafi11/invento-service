package helper

import (
	"encoding/base64"
	"strings"
)

func ParseTusMetadata(header string) map[string]string {
	metadata := make(map[string]string)
	if header == "" {
		return metadata
	}

	pairs := strings.Split(header, ",")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		parts := strings.SplitN(pair, " ", 2)
		if len(parts) == 0 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		if key == "" {
			continue
		}

		value := ""
		if len(parts) == 2 {
			decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(parts[1]))
			if err == nil {
				value = string(decoded)
			}
		}

		metadata[key] = value
	}

	return metadata
}
