package main

// #include <stdbool.h>
// #include "import.h"
import "C"
import (
	"net/url"
	"strings"
	"unsafe"

	"github.com/bakape/hydron/core"
	"github.com/mailru/easyjson/jwriter"
)

// Wraps function pointers to implement core.FetchLogger
type ffiLogger struct {
	passOnDone                    bool // Pass records as JSON, when done
	total, done                   int
	setState, passDone, passError unsafe.Pointer
}

func (l *ffiLogger) SetTotal(n int) {
	l.total = n
	l.passState()
}

func (l *ffiLogger) passState() {
	if l.setState == nil {
		return
	}
	var pos float64
	if l.total != 0 {
		pos = float64(l.done) / float64(l.total)
	}
	C.call_float_func(C.float_func(l.setState), C.double(pos))
}

func (l *ffiLogger) Done(kv core.KeyValue) {
	l.done++
	l.passState()

	// kv is empty, if file is alsready imported
	if l.passOnDone && kv.SHA1 != [20]byte{} {
		go func() {
			var w jwriter.Writer
			encodeRecord(kv, &w, true)
			rec, err := toCJSON(&w)
			if err != nil {
				C.call_char_func(C.char_func(l.passError), err)
			} else {
				C.call_char_func(C.char_func(l.passDone), rec)
			}
		}()
	}
}

func (l *ffiLogger) Err(err error) {
	C.call_char_func(C.char_func(l.passError), toCError(err))
}

func (l *ffiLogger) Close() {}

func (l *ffiLogger) NoMatch(kv core.KeyValue) {
	// TODO: Fuzy match against IQDB
	l.Done(kv)
}

//export importFiles
func importFiles(
	files, tags *C.char,
	deleteAfter, fetch C.bool,
	setState unsafe.Pointer,
	passDone unsafe.Pointer,
	passError unsafe.Pointer,
) {
	go func() {
		iLog := ffiLogger{
			passOnDone: !bool(fetch),
			setState:   setState,
			passDone:   passDone,
			passError:  passError,
		}
		fLog := ffiLogger{
			passOnDone: true,
			passDone:   passDone,
			passError:  passError,
		}

		paths := strings.Split(C.GoString(files), "\n")
		var err error
		for i, p := range paths {
			paths[i], err = url.PathUnescape(strings.TrimPrefix(p, "file://"))
			if err != nil {
				iLog.Err(err)
				return
			}
		}

		var parsedTags [][]byte
		if tags != nil {
			parsedTags = core.SplitTagString(C.GoString(tags), ' ')
		}

		err = core.ImportPaths(
			paths,
			bool(deleteAfter), bool(fetch),
			parsedTags,
			&iLog, &fLog,
		)
		if err != nil {
			iLog.Err(err)
		}

		// Reset progress bar
		iLog.total = 0
		iLog.passState()
	}()
}
