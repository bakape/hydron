package core

import (
	"bytes"
	"fmt"
	"strconv"
)

// Types of system values to retrieve
const (
	sizeValue = iota
	widthValue
	heightValue
	lengthValue
	tagCountValue
)

// Used for autocompletion
var systemTagHeaders = []string{
	"size", "width", "height", "length", "tag_count",
}

type syntaxError []byte

func (e syntaxError) Error() string {
	return fmt.Sprintf("syntax error: %s", string(e))
}

type systemTagTest struct {
	typ  uint8
	comp byte
	val  uint64
}

func isSystemTag(t []byte) bool {
	return bytes.HasPrefix(t, []byte("system:"))
}

// Further search for matches by internal system tags like size and width.
// If matched is nil, the entire database is searched.
func searchBySystemTags(matched map[[20]byte]bool, tags [][]byte) (
	rematched map[[20]byte]bool, err error,
) {
	tests, err := parseSystemTags(tags)
	if err != nil {
		return
	}

	check := func(k []byte, r Record) {
		pass := true

		for _, t := range tests {
			// Extract values from Record
			var val uint64
			switch t.typ {
			case sizeValue:
				val = r.Size()
			case widthValue:
				val = r.Width()
			case heightValue:
				val = r.Height()
			case lengthValue:
				val = r.Length()
			case tagCountValue:
				val = uint64(len(r.Tags()))
			}

			// Compare to test argument
			switch t.comp {
			case '=':
				pass = val == t.val
			case '>':
				pass = val > t.val
			case '<':
				pass = val < t.val
			}
			if !pass {
				break
			}
		}

		if !pass {
			// If matched is not nil, remove not passing values from an
			// existing set
			if matched != nil {
				delete(rematched, ExtractKey(k))
			}
		} else {
			// If matched is nil, add passed values to a fresh new set
			if matched == nil {
				rematched[ExtractKey(k)] = true
			}
		}
	}

	// Search entire database
	if matched == nil {
		rematched = make(map[[20]byte]bool, 128)
		err = IterateRecords(check)
		return
	}

	// Search only the already matched keys
	rematched = matched
	tx, err := db.Begin(false)
	if err != nil {
		return
	}

	buc := tx.Bucket([]byte("images"))
	for id := range matched {
		check(id[:], Record(buc.Get(id[:])))
	}
	err = tx.Rollback()

	return
}

func parseSystemTags(tags [][]byte) (tests []systemTagTest, err error) {
	tests = make([]systemTagTest, 0, len(tags))

	for i, t := range tags {
		var test systemTagTest
		t = t[7:]

	parser:
		for j, b := range t {
			// Parse value type and comparison operator
			switch {
			case ('a' <= b && b <= 'z') || b == '_':
			case b == '>' || b == '<' || b == '=':
				// String too short
				if j+1 >= len(t) {
					err = syntaxError(tags[i])
					return
				}

				switch string(t[:j]) {
				case "size":
					test.typ = sizeValue
				case "width":
					test.typ = widthValue
				case "height":
					test.typ = heightValue
				case "length":
					test.typ = lengthValue
				case "tag_count":
					test.typ = tagCountValue
				default:
					err = syntaxError(tags[i])
					return
				}

				test.comp = b

				// Parse numeric value to compare to
				test.val, err = strconv.ParseUint(string(t[j+1:]), 10, 64)
				if err != nil {
					err = syntaxError(string(tags[i]) + " : " + err.Error())
					return
				}

				tests = append(tests, test)
				break parser
			default:
				err = syntaxError(tags[i])
				return
			}
		}
	}

	return
}
