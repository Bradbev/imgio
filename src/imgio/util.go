package imgio

import (
	"regexp"
	"strings"
)

var findHashes = regexp.MustCompile("(.*?)(##.*)").FindStringSubmatch

// From a string, extract the id that should be used uniquely
// Follows imgui's label conventions of ## and ###
func getId(str, idType string) (label, id string) {
	parts := findHashes(str)
	if len(parts) == 0 {
		return str, str
	}
	if strings.HasPrefix(parts[2], "###") {
		return parts[1], parts[2]
	}
	return parts[1], parts[0]
}
