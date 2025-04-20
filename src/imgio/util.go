package imgio

import (
	"regexp"
	"strings"

	"gioui.org/io/event"
	"gioui.org/io/input"
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

// forEvent will filter s according to filter and try to convert any matching events
// to type T.  body will be called for all events that pass both the filter and the cast.
// To exit early, body can return false
func forEvent[T any](s input.Source, filter event.Filter, body func(evt T) bool) {
	for {
		ev, ok := s.Event(filter)
		if !ok {
			return
		}
		e, ok := ev.(T)
		if !ok {
			continue
		}
		if !body(e) {
			return
		}
	}
}

func fromCache[T any](i *Im, key string, makeValue func() T) T {
	item, exists := i.widgets[key]
	if !exists {
		item = makeValue()
		i.widgets[key] = item
	}
	return item.(T)
}
