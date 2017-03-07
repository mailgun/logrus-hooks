// AUTOGENERATED FILE: easyjson marshaler/unmarshalers.

package udploghook

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

func easyjson74ee6e88DecodeGithubComMailgunLogrusUdplog(in *jlexer.Lexer, out *logRecord) {
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
		case "context":
			if in.IsNull() {
				in.Skip()
			} else {
				in.Delim('{')
				if !in.IsDelim('}') {
					out.Context = make(map[string]interface{})
				} else {
					out.Context = nil
				}
				for !in.IsDelim('}') {
					key := string(in.String())
					in.WantColon()
					var v1 interface{}
					if m, ok := v1.(easyjson.Unmarshaler); ok {
						m.UnmarshalEasyJSON(in)
					} else if m, ok := v1.(json.Unmarshaler); ok {
						m.UnmarshalJSON(in.Raw())
					} else {
						v1 = in.Interface()
					}
					(out.Context)[key] = v1
					in.WantComma()
				}
				in.Delim('}')
			}
		case "appname":
			out.AppName = string(in.String())
		case "hostname":
			out.HostName = string(in.String())
		case "logLevel":
			out.LogLevel = string(in.String())
		case "filename":
			out.FileName = string(in.String())
		case "funcName":
			out.FuncName = string(in.String())
		case "lineno":
			out.LineNo = int(in.Int())
		case "message":
			out.Message = string(in.String())
		case "timestamp":
			out.Timestamp = Number(in.Float64())
		case "pid":
			out.PID = int(in.Int())
		case "tid":
			out.TID = string(in.String())
		case "excType":
			out.ExcType = string(in.String())
		case "excText":
			out.ExcText = string(in.String())
		case "excValue":
			out.ExcValue = string(in.String())
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
func easyjson74ee6e88EncodeGithubComMailgunLogrusUdplog(out *jwriter.Writer, in logRecord) {
	out.RawByte('{')
	first := true
	_ = first
	if len(in.Context) != 0 {
		if !first {
			out.RawByte(',')
		}
		first = false
		out.RawString("\"context\":")
		if in.Context == nil && (out.Flags&jwriter.NilMapAsEmpty) == 0 {
			out.RawString(`null`)
		} else {
			out.RawByte('{')
			v2First := true
			for v2Name, v2Value := range in.Context {
				if !v2First {
					out.RawByte(',')
				}
				v2First = false
				out.String(string(v2Name))
				out.RawByte(':')
				if m, ok := v2Value.(easyjson.Marshaler); ok {
					m.MarshalEasyJSON(out)
				} else if m, ok := v2Value.(json.Marshaler); ok {
					out.Raw(m.MarshalJSON())
				} else {
					out.Raw(json.Marshal(v2Value))
				}
			}
			out.RawByte('}')
		}
	}
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"appname\":")
	out.String(string(in.AppName))
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"hostname\":")
	out.String(string(in.HostName))
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"logLevel\":")
	out.String(string(in.LogLevel))
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"filename\":")
	out.String(string(in.FileName))
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"funcName\":")
	out.String(string(in.FuncName))
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"lineno\":")
	out.Int(int(in.LineNo))
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"message\":")
	out.String(string(in.Message))
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"timestamp\":")
	out.Raw((in.Timestamp).MarshalJSON())
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"pid\":")
	out.Int(int(in.PID))
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"tid\":")
	out.String(string(in.TID))
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"excType\":")
	out.String(string(in.ExcType))
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"excText\":")
	out.String(string(in.ExcText))
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"excValue\":")
	out.String(string(in.ExcValue))
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v logRecord) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson74ee6e88EncodeGithubComMailgunLogrusUdplog(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v logRecord) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson74ee6e88EncodeGithubComMailgunLogrusUdplog(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *logRecord) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson74ee6e88DecodeGithubComMailgunLogrusUdplog(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *logRecord) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson74ee6e88DecodeGithubComMailgunLogrusUdplog(l, v)
}
