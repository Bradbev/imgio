package imgio

import (
	"math"
	"regexp"
	"strings"

	"gioui.org/f32"
	"gioui.org/io/event"
	"gioui.org/io/input"
	"golang.org/x/exp/constraints"
)

var findHashes = regexp.MustCompile("(.*?)(##.*)").FindStringSubmatch

// From a string, extract the id that should be used uniquely
// Follows imgui's label conventions of ## and ###
func getId(str, idType string) (label, id string) {
	parts := findHashes(str)
	if len(parts) == 0 {
		return str, str + idType
	}
	if strings.HasPrefix(parts[2], "###") {
		return parts[1], parts[2] + idType
	}
	return parts[1], parts[0] + idType
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

func trunc32(f float32) float32 {
	return float32(math.Trunc(float64(f)))
}

func truncPt(p f32.Point) f32.Point {
	return f32.Pt(trunc32(p.X), trunc32(p.Y))
}

func clamp[T constraints.Float | constraints.Integer](v, min, max T) T {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
