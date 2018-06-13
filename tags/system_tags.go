package tags

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Types of system values to retrieve
type SystemTagType uint8

const (
	Size SystemTagType = iota
	Width
	Height
	Duration
	TagCount
)

// Parsed data of system tag
type SystemTag struct {
	Type       SystemTagType
	Comparator string
	Value      uint64
}

// Type of ordering for search results
type Ordering struct {
	Type    OrderingType
	Reverse bool
}

// Types of ordering for search results
type OrderingType uint8

const (
	None OrderingType = iota
	BySize
	ByWidth
	ByHeight
	ByDuration
	ByTagCount
	Random
)

var (
	// Possible headers of internal tag matching rules
	SystemHeaders = []string{"size", "width", "height", "duration", "tag_count"}
	systemRegex   = regexp.MustCompile(`^([\w_]+)(=|>|>=|<|<=)(\d+)$`)
)

type SyntaxError string

func (e SyntaxError) Error() string {
	return fmt.Sprintf("syntax error: %s", string(e))
}

// Extract system, ordering and limit parameters from tag list
func ExtractInternal(tags *[]string) (
	sys []SystemTag, orderBy Ordering, limit uint64, err error,
) {
	mod := make([]string, 0, len(*tags))
	for _, t := range *tags {
		t = strings.TrimSpace(t)
		i := strings.IndexByte(t, ':')
		if i != -1 {
			arg := t[i+1:]
			switch t[:i] {
			case "system":
				var s SystemTag
				s, err = parseSystemTag(arg)
				sys = append(sys, s)
			case "order":
				orderBy, err = parseOrdering(arg)
			case "limit":
				limit, err = strconv.ParseUint(arg, 10, 64)
			default:
				goto unmatched
			}
			if err != nil {
				return
			}
			continue
		}
	unmatched:
		if len(t) != 0 { // Might as well filter empty spaces here
			mod = append(mod, t)
		}
	}

	*tags = mod
	return
}

func parseSystemTag(arg string) (sys SystemTag, err error) {
	m := systemRegex.FindStringSubmatch(arg)
	if m == nil {
		err = SyntaxError(arg)
		return
	}

	switch m[1] {
	case "size":
		sys.Type = Size
	case "width":
		sys.Type = Width
	case "height":
		sys.Type = Height
	case "duration":
		sys.Type = Duration
	case "tag_count":
		sys.Type = TagCount
	default:
		err = SyntaxError(arg)
		return
	}
	sys.Comparator = m[2]
	sys.Value, err = strconv.ParseUint(m[3], 10, 64)
	if err != nil {
		err = SyntaxError(err.Error())
	}

	return
}

func parseOrdering(arg string) (o Ordering, err error) {
	if len(arg) < 4 {
		err = SyntaxError(arg)
		return
	}
	if arg[0] == '-' {
		o.Reverse = true
		arg = arg[1:]
	}
	switch arg {
	case "size":
		o.Type = BySize
	case "width":
		o.Type = ByWidth
	case "height":
		o.Type = ByHeight
	case "duration":
		o.Type = ByDuration
	case "tag_count":
		o.Type = ByTagCount
	case "random":
		o.Type = Random
	default:
		err = SyntaxError(arg)
	}
	return
}
