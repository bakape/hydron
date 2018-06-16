// This file is automatically generated by qtc from "util.qtpl".
// See https://github.com/valyala/quicktemplate for details.

//line util.qtpl:1
package templates

//line util.qtpl:1
import (
	qtio422016 "io"

	qt422016 "github.com/valyala/quicktemplate"
)

//line util.qtpl:1
var (
	_ = qtio422016.Copy
	_ = qt422016.AcquireByteBuffer
)

//line util.qtpl:1
func streamhead(qw422016 *qt422016.Writer, title string) {
	//line util.qtpl:1
	qw422016.N().S(`<!doctype html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, minimum-scale=1.0, maximum-scale=1.0"><title>`)
	//line util.qtpl:6
	qw422016.E().S(title)
	//line util.qtpl:6
	qw422016.N().S(`</title><link rel="stylesheet" href="/assets/main.css"></head>`)
//line util.qtpl:9
}

//line util.qtpl:9
func writehead(qq422016 qtio422016.Writer, title string) {
	//line util.qtpl:9
	qw422016 := qt422016.AcquireWriter(qq422016)
	//line util.qtpl:9
	streamhead(qw422016, title)
	//line util.qtpl:9
	qt422016.ReleaseWriter(qw422016)
//line util.qtpl:9
}

//line util.qtpl:9
func head(title string) string {
	//line util.qtpl:9
	qb422016 := qt422016.AcquireByteBuffer()
	//line util.qtpl:9
	writehead(qb422016, title)
	//line util.qtpl:9
	qs422016 := string(qb422016.B)
	//line util.qtpl:9
	qt422016.ReleaseByteBuffer(qb422016)
	//line util.qtpl:9
	return qs422016
//line util.qtpl:9
}
