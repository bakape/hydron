// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package common

import (
	json "encoding/json"
	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
)

// suppress unused package warning
var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

func easyjson15d5d517DecodeGithubComBakapeHydronCommon(in *jlexer.Lexer, out *Thumbnail) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "is_png":
			out.IsPNG = bool(in.Bool())
		case "width":
			out.Width = uint64(in.Uint64())
		case "height":
			out.Height = uint64(in.Uint64())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson15d5d517EncodeGithubComBakapeHydronCommon(out *jwriter.Writer, in Thumbnail) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"is_png\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Bool(bool(in.IsPNG))
	}
	{
		const prefix string = ",\"width\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Uint64(uint64(in.Width))
	}
	{
		const prefix string = ",\"height\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Uint64(uint64(in.Height))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v Thumbnail) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson15d5d517EncodeGithubComBakapeHydronCommon(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Thumbnail) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson15d5d517EncodeGithubComBakapeHydronCommon(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *Thumbnail) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson15d5d517DecodeGithubComBakapeHydronCommon(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Thumbnail) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson15d5d517DecodeGithubComBakapeHydronCommon(l, v)
}
func easyjson15d5d517DecodeGithubComBakapeHydronCommon1(in *jlexer.Lexer, out *Record) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "type":
			out.Type = FileType(in.Uint8())
		case "importTime":
			out.ImportTime = uint64(in.Uint64())
		case "size":
			out.Size = uint64(in.Uint64())
		case "duration":
			out.Duration = uint64(in.Uint64())
		case "md5":
			out.MD5 = string(in.String())
		case "tags":
			if in.IsNull() {
				in.Skip()
				out.Tags = nil
			} else {
				in.Delim('[')
				if out.Tags == nil {
					if !in.IsDelim(']') {
						out.Tags = make([]Tag, 0, 2)
					} else {
						out.Tags = []Tag{}
					}
				} else {
					out.Tags = (out.Tags)[:0]
				}
				for !in.IsDelim(']') {
					var v1 Tag
					easyjson15d5d517DecodeGithubComBakapeHydronCommon2(in, &v1)
					out.Tags = append(out.Tags, v1)
					in.WantComma()
				}
				in.Delim(']')
			}
		case "width":
			out.Width = uint64(in.Uint64())
		case "height":
			out.Height = uint64(in.Uint64())
		case "sha1":
			out.SHA1 = string(in.String())
		case "thumb":
			(out.Thumb).UnmarshalEasyJSON(in)
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson15d5d517EncodeGithubComBakapeHydronCommon1(out *jwriter.Writer, in Record) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"type\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Uint8(uint8(in.Type))
	}
	{
		const prefix string = ",\"importTime\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Uint64(uint64(in.ImportTime))
	}
	{
		const prefix string = ",\"size\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Uint64(uint64(in.Size))
	}
	if in.Duration != 0 {
		const prefix string = ",\"duration\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Uint64(uint64(in.Duration))
	}
	{
		const prefix string = ",\"md5\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.MD5))
	}
	if len(in.Tags) != 0 {
		const prefix string = ",\"tags\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		{
			out.RawByte('[')
			for v2, v3 := range in.Tags {
				if v2 > 0 {
					out.RawByte(',')
				}
				easyjson15d5d517EncodeGithubComBakapeHydronCommon2(out, v3)
			}
			out.RawByte(']')
		}
	}
	{
		const prefix string = ",\"width\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Uint64(uint64(in.Width))
	}
	{
		const prefix string = ",\"height\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Uint64(uint64(in.Height))
	}
	{
		const prefix string = ",\"sha1\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.SHA1))
	}
	{
		const prefix string = ",\"thumb\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		(in.Thumb).MarshalEasyJSON(out)
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v Record) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson15d5d517EncodeGithubComBakapeHydronCommon1(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Record) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson15d5d517EncodeGithubComBakapeHydronCommon1(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *Record) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson15d5d517DecodeGithubComBakapeHydronCommon1(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Record) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson15d5d517DecodeGithubComBakapeHydronCommon1(l, v)
}
func easyjson15d5d517DecodeGithubComBakapeHydronCommon2(in *jlexer.Lexer, out *Tag) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "type":
			out.Type = TagType(in.Uint8())
		case "source":
			out.Source = TagSource(in.Uint8())
		case "tag":
			out.Tag = string(in.String())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson15d5d517EncodeGithubComBakapeHydronCommon2(out *jwriter.Writer, in Tag) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"type\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Uint8(uint8(in.Type))
	}
	{
		const prefix string = ",\"source\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Uint8(uint8(in.Source))
	}
	{
		const prefix string = ",\"tag\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.Tag))
	}
	out.RawByte('}')
}
func easyjson15d5d517DecodeGithubComBakapeHydronCommon3(in *jlexer.Lexer, out *Image) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "width":
			out.Width = uint64(in.Uint64())
		case "height":
			out.Height = uint64(in.Uint64())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson15d5d517EncodeGithubComBakapeHydronCommon3(out *jwriter.Writer, in Image) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"width\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Uint64(uint64(in.Width))
	}
	{
		const prefix string = ",\"height\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Uint64(uint64(in.Height))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v Image) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson15d5d517EncodeGithubComBakapeHydronCommon3(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Image) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson15d5d517EncodeGithubComBakapeHydronCommon3(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *Image) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson15d5d517DecodeGithubComBakapeHydronCommon3(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Image) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson15d5d517DecodeGithubComBakapeHydronCommon3(l, v)
}
func easyjson15d5d517DecodeGithubComBakapeHydronCommon4(in *jlexer.Lexer, out *CompactRecord) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "sha1":
			out.SHA1 = string(in.String())
		case "thumb":
			(out.Thumb).UnmarshalEasyJSON(in)
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson15d5d517EncodeGithubComBakapeHydronCommon4(out *jwriter.Writer, in CompactRecord) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"sha1\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.SHA1))
	}
	{
		const prefix string = ",\"thumb\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		(in.Thumb).MarshalEasyJSON(out)
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v CompactRecord) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson15d5d517EncodeGithubComBakapeHydronCommon4(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v CompactRecord) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson15d5d517EncodeGithubComBakapeHydronCommon4(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *CompactRecord) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson15d5d517DecodeGithubComBakapeHydronCommon4(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *CompactRecord) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson15d5d517DecodeGithubComBakapeHydronCommon4(l, v)
}
