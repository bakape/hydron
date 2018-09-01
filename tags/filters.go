package tags

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/bakape/hydron/common"
)

var systemRegex = regexp.MustCompile(`^([\w_]+)(=|>|>=|<|<=)(\d+)$`)

type SyntaxError string

func (e SyntaxError) Error() string {
	return fmt.Sprintf("syntax error: %s", string(e))
}

// Implement main.StatusError
func (e SyntaxError) Status() int {
	return 400
}

// Parse string into filter list and extract system, ordering and limit
// parameters from tag list
func ParseFilters(query string, page *common.Page) (err error) {
	for _, t := range strings.Split(query, " ") {
		if len(t) == 0 {
			continue
		}

		i := strings.IndexByte(t, ':')
		if i != -1 {
			arg := t[i+1:]
			switch t[:i] {
			case "system":
				var s common.SystemTag
				s, err = parseSystemTag(arg)
				page.Filters.System = append(page.Filters.System, s)
			case "order":
				err = parseOrdering(arg, &page.Order)
			case "limit":
				var j uint64
				j, err = strconv.ParseUint(arg, 10, 64)
				page.Limit = uint(j)
			default:
				goto normalTag
			}
			if err != nil {
				return
			}
			continue
		}

	normalTag:
		page.Filters.Tag = append(page.Filters.Tag, common.TagFilter{
			Negative: isNegative(&t),
			TagBase: common.TagBase{
				Type: detectTagType(&t),
				Tag:  normalizeString(t),
			},
		})
	}

	return
}

func parseSystemTag(arg string) (sys common.SystemTag, err error) {
	m := systemRegex.FindStringSubmatch(arg)
	if m == nil {
		err = SyntaxError(arg)
		return
	}

	switch m[1] {
	case "size":
		sys.Type = common.Size
	case "width":
		sys.Type = common.Width
	case "height":
		sys.Type = common.Height
	case "duration":
		sys.Type = common.Duration
	case "tag_count":
		sys.Type = common.TagCount
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

func isNegative(s *string) bool {
	if (*s)[0] == '-' {
		*s = (*s)[1:]
		return true
	}
	return false
}

func parseOrdering(arg string, o *common.Order) error {
	if len(arg) < 4 {
		return SyntaxError(arg)
	}
	o.Reverse = isNegative(&arg)
	switch arg {
	case "size":
		o.Type = common.BySize
	case "width":
		o.Type = common.ByWidth
	case "height":
		o.Type = common.ByHeight
	case "duration":
		o.Type = common.ByDuration
	case "tag_count":
		o.Type = common.ByTagCount
	case "random":
		o.Type = common.Random
	default:
		return SyntaxError(arg)
	}
	return nil
}
