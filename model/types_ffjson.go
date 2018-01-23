// Code generated by ffjson <https://github.com/pquerna/ffjson>. DO NOT EDIT.
// source: model/types.go

package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	fflib "github.com/pquerna/ffjson/fflib/v1"
)

const (
	ffjtJsonRsyslogMessagebase = iota
	ffjtJsonRsyslogMessagenosuchkey

	ffjtJsonRsyslogMessageMessage

	ffjtJsonRsyslogMessageTimeReported

	ffjtJsonRsyslogMessageTimeGenerated

	ffjtJsonRsyslogMessageHostname

	ffjtJsonRsyslogMessagePriority

	ffjtJsonRsyslogMessageAppname

	ffjtJsonRsyslogMessageProcid

	ffjtJsonRsyslogMessageMsgid

	ffjtJsonRsyslogMessageUuid

	ffjtJsonRsyslogMessageStructured

	ffjtJsonRsyslogMessageProperties
)

var ffjKeyJsonRsyslogMessageMessage = []byte("msg")

var ffjKeyJsonRsyslogMessageTimeReported = []byte("timereported")

var ffjKeyJsonRsyslogMessageTimeGenerated = []byte("timegenerated")

var ffjKeyJsonRsyslogMessageHostname = []byte("hostname")

var ffjKeyJsonRsyslogMessagePriority = []byte("pri")

var ffjKeyJsonRsyslogMessageAppname = []byte("app-name")

var ffjKeyJsonRsyslogMessageProcid = []byte("procid")

var ffjKeyJsonRsyslogMessageMsgid = []byte("msgid")

var ffjKeyJsonRsyslogMessageUuid = []byte("uuid")

var ffjKeyJsonRsyslogMessageStructured = []byte("structured-data")

var ffjKeyJsonRsyslogMessageProperties = []byte("$!")

// UnmarshalJSON umarshall json - template of ffjson
func (j *JsonRsyslogMessage) UnmarshalJSON(input []byte) error {
	fs := fflib.NewFFLexer(input)
	return j.UnmarshalJSONFFLexer(fs, fflib.FFParse_map_start)
}

// UnmarshalJSONFFLexer fast json unmarshall - template ffjson
func (j *JsonRsyslogMessage) UnmarshalJSONFFLexer(fs *fflib.FFLexer, state fflib.FFParseState) error {
	var err error
	currentKey := ffjtJsonRsyslogMessagebase
	_ = currentKey
	tok := fflib.FFTok_init
	wantedTok := fflib.FFTok_init

mainparse:
	for {
		tok = fs.Scan()
		//	println(fmt.Sprintf("debug: tok: %v  state: %v", tok, state))
		if tok == fflib.FFTok_error {
			goto tokerror
		}

		switch state {

		case fflib.FFParse_map_start:
			if tok != fflib.FFTok_left_bracket {
				wantedTok = fflib.FFTok_left_bracket
				goto wrongtokenerror
			}
			state = fflib.FFParse_want_key
			continue

		case fflib.FFParse_after_value:
			if tok == fflib.FFTok_comma {
				state = fflib.FFParse_want_key
			} else if tok == fflib.FFTok_right_bracket {
				goto done
			} else {
				wantedTok = fflib.FFTok_comma
				goto wrongtokenerror
			}

		case fflib.FFParse_want_key:
			// json {} ended. goto exit. woo.
			if tok == fflib.FFTok_right_bracket {
				goto done
			}
			if tok != fflib.FFTok_string {
				wantedTok = fflib.FFTok_string
				goto wrongtokenerror
			}

			kn := fs.Output.Bytes()
			if len(kn) <= 0 {
				// "" case. hrm.
				currentKey = ffjtJsonRsyslogMessagenosuchkey
				state = fflib.FFParse_want_colon
				goto mainparse
			} else {
				switch kn[0] {

				case '$':

					if bytes.Equal(ffjKeyJsonRsyslogMessageProperties, kn) {
						currentKey = ffjtJsonRsyslogMessageProperties
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				case 'a':

					if bytes.Equal(ffjKeyJsonRsyslogMessageAppname, kn) {
						currentKey = ffjtJsonRsyslogMessageAppname
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				case 'h':

					if bytes.Equal(ffjKeyJsonRsyslogMessageHostname, kn) {
						currentKey = ffjtJsonRsyslogMessageHostname
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				case 'm':

					if bytes.Equal(ffjKeyJsonRsyslogMessageMessage, kn) {
						currentKey = ffjtJsonRsyslogMessageMessage
						state = fflib.FFParse_want_colon
						goto mainparse

					} else if bytes.Equal(ffjKeyJsonRsyslogMessageMsgid, kn) {
						currentKey = ffjtJsonRsyslogMessageMsgid
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				case 'p':

					if bytes.Equal(ffjKeyJsonRsyslogMessagePriority, kn) {
						currentKey = ffjtJsonRsyslogMessagePriority
						state = fflib.FFParse_want_colon
						goto mainparse

					} else if bytes.Equal(ffjKeyJsonRsyslogMessageProcid, kn) {
						currentKey = ffjtJsonRsyslogMessageProcid
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				case 's':

					if bytes.Equal(ffjKeyJsonRsyslogMessageStructured, kn) {
						currentKey = ffjtJsonRsyslogMessageStructured
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				case 't':

					if bytes.Equal(ffjKeyJsonRsyslogMessageTimeReported, kn) {
						currentKey = ffjtJsonRsyslogMessageTimeReported
						state = fflib.FFParse_want_colon
						goto mainparse

					} else if bytes.Equal(ffjKeyJsonRsyslogMessageTimeGenerated, kn) {
						currentKey = ffjtJsonRsyslogMessageTimeGenerated
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				case 'u':

					if bytes.Equal(ffjKeyJsonRsyslogMessageUuid, kn) {
						currentKey = ffjtJsonRsyslogMessageUuid
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				}

				if fflib.AsciiEqualFold(ffjKeyJsonRsyslogMessageProperties, kn) {
					currentKey = ffjtJsonRsyslogMessageProperties
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.EqualFoldRight(ffjKeyJsonRsyslogMessageStructured, kn) {
					currentKey = ffjtJsonRsyslogMessageStructured
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.SimpleLetterEqualFold(ffjKeyJsonRsyslogMessageUuid, kn) {
					currentKey = ffjtJsonRsyslogMessageUuid
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.EqualFoldRight(ffjKeyJsonRsyslogMessageMsgid, kn) {
					currentKey = ffjtJsonRsyslogMessageMsgid
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.SimpleLetterEqualFold(ffjKeyJsonRsyslogMessageProcid, kn) {
					currentKey = ffjtJsonRsyslogMessageProcid
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.AsciiEqualFold(ffjKeyJsonRsyslogMessageAppname, kn) {
					currentKey = ffjtJsonRsyslogMessageAppname
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.SimpleLetterEqualFold(ffjKeyJsonRsyslogMessagePriority, kn) {
					currentKey = ffjtJsonRsyslogMessagePriority
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.EqualFoldRight(ffjKeyJsonRsyslogMessageHostname, kn) {
					currentKey = ffjtJsonRsyslogMessageHostname
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.SimpleLetterEqualFold(ffjKeyJsonRsyslogMessageTimeGenerated, kn) {
					currentKey = ffjtJsonRsyslogMessageTimeGenerated
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.SimpleLetterEqualFold(ffjKeyJsonRsyslogMessageTimeReported, kn) {
					currentKey = ffjtJsonRsyslogMessageTimeReported
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.EqualFoldRight(ffjKeyJsonRsyslogMessageMessage, kn) {
					currentKey = ffjtJsonRsyslogMessageMessage
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				currentKey = ffjtJsonRsyslogMessagenosuchkey
				state = fflib.FFParse_want_colon
				goto mainparse
			}

		case fflib.FFParse_want_colon:
			if tok != fflib.FFTok_colon {
				wantedTok = fflib.FFTok_colon
				goto wrongtokenerror
			}
			state = fflib.FFParse_want_value
			continue
		case fflib.FFParse_want_value:

			if tok == fflib.FFTok_left_brace || tok == fflib.FFTok_left_bracket || tok == fflib.FFTok_integer || tok == fflib.FFTok_double || tok == fflib.FFTok_string || tok == fflib.FFTok_bool || tok == fflib.FFTok_null {
				switch currentKey {

				case ffjtJsonRsyslogMessageMessage:
					goto handle_Message

				case ffjtJsonRsyslogMessageTimeReported:
					goto handle_TimeReported

				case ffjtJsonRsyslogMessageTimeGenerated:
					goto handle_TimeGenerated

				case ffjtJsonRsyslogMessageHostname:
					goto handle_Hostname

				case ffjtJsonRsyslogMessagePriority:
					goto handle_Priority

				case ffjtJsonRsyslogMessageAppname:
					goto handle_Appname

				case ffjtJsonRsyslogMessageProcid:
					goto handle_Procid

				case ffjtJsonRsyslogMessageMsgid:
					goto handle_Msgid

				case ffjtJsonRsyslogMessageUuid:
					goto handle_Uuid

				case ffjtJsonRsyslogMessageStructured:
					goto handle_Structured

				case ffjtJsonRsyslogMessageProperties:
					goto handle_Properties

				case ffjtJsonRsyslogMessagenosuchkey:
					err = fs.SkipField(tok)
					if err != nil {
						return fs.WrapErr(err)
					}
					state = fflib.FFParse_after_value
					goto mainparse
				}
			} else {
				goto wantedvalue
			}
		}
	}

handle_Message:

	/* handler: j.Message type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			j.Message = string(string(outBuf))

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_TimeReported:

	/* handler: j.TimeReported type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			j.TimeReported = string(string(outBuf))

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_TimeGenerated:

	/* handler: j.TimeGenerated type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			j.TimeGenerated = string(string(outBuf))

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_Hostname:

	/* handler: j.Hostname type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			j.Hostname = string(string(outBuf))

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_Priority:

	/* handler: j.Priority type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			j.Priority = string(string(outBuf))

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_Appname:

	/* handler: j.Appname type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			j.Appname = string(string(outBuf))

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_Procid:

	/* handler: j.Procid type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			j.Procid = string(string(outBuf))

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_Msgid:

	/* handler: j.Msgid type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			j.Msgid = string(string(outBuf))

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_Uuid:

	/* handler: j.Uuid type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			j.Uuid = string(string(outBuf))

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_Structured:

	/* handler: j.Structured type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			j.Structured = string(string(outBuf))

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_Properties:

	/* handler: j.Properties type=map[string]interface {} kind=map quoted=false*/

	{

		{
			if tok != fflib.FFTok_left_bracket && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for ", tok))
			}
		}

		if tok == fflib.FFTok_null {
			j.Properties = nil
		} else {

			j.Properties = make(map[string]interface{}, 0)

			wantVal := true

			for {

				var k string

				var tmpJProperties interface{}

				tok = fs.Scan()
				if tok == fflib.FFTok_error {
					goto tokerror
				}
				if tok == fflib.FFTok_right_bracket {
					break
				}

				if tok == fflib.FFTok_comma {
					if wantVal == true {
						// TODO(pquerna): this isn't an ideal error message, this handles
						// things like [,,,] as an array value.
						return fs.WrapErr(fmt.Errorf("wanted value token, but got token: %v", tok))
					}
					continue
				} else {
					wantVal = true
				}

				/* handler: k type=string kind=string quoted=false*/

				{

					{
						if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
							return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
						}
					}

					if tok == fflib.FFTok_null {

					} else {

						outBuf := fs.Output.Bytes()

						k = string(string(outBuf))

					}
				}

				// Expect ':' after key
				tok = fs.Scan()
				if tok != fflib.FFTok_colon {
					return fs.WrapErr(fmt.Errorf("wanted colon token, but got token: %v", tok))
				}

				tok = fs.Scan()
				/* handler: tmpJProperties type=interface {} kind=interface quoted=false*/

				{
					/* Falling back. type=interface {} kind=interface */
					tbuf, err := fs.CaptureField(tok)
					if err != nil {
						return fs.WrapErr(err)
					}

					err = json.Unmarshal(tbuf, &tmpJProperties)
					if err != nil {
						return fs.WrapErr(err)
					}
				}

				j.Properties[k] = tmpJProperties

				wantVal = false
			}

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

wantedvalue:
	return fs.WrapErr(fmt.Errorf("wanted value token, but got token: %v", tok))
wrongtokenerror:
	return fs.WrapErr(fmt.Errorf("ffjson: wanted token: %v, but got token: %v output=%s", wantedTok, tok, fs.Output.String()))
tokerror:
	if fs.BigError != nil {
		return fs.WrapErr(fs.BigError)
	}
	err = fs.Error.ToError()
	if err != nil {
		return fs.WrapErr(err)
	}
	panic("ffjson-generated: unreachable, please report bug.")
done:

	return nil
}

// MarshalJSON marshal bytes to json - template
func (j *RegularSyslog) MarshalJSON() ([]byte, error) {
	var buf fflib.Buffer
	if j == nil {
		buf.WriteString("null")
		return buf.Bytes(), nil
	}
	err := j.MarshalJSONBuf(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// MarshalJSONBuf marshal buff to json - template
func (j *RegularSyslog) MarshalJSONBuf(buf fflib.EncodingBuffer) error {
	if j == nil {
		buf.WriteString("null")
		return nil
	}
	var err error
	var obj []byte
	_ = obj
	_ = err
	buf.WriteString(`{"facility":`)
	fflib.WriteJsonString(buf, string(j.Facility))
	buf.WriteString(`,"severity":`)
	fflib.WriteJsonString(buf, string(j.Severity))
	buf.WriteString(`,"timereported":`)

	{

		obj, err = j.TimeReported.MarshalJSON()
		if err != nil {
			return err
		}
		buf.Write(obj)

	}
	buf.WriteString(`,"timegenerated":`)

	{

		obj, err = j.TimeGenerated.MarshalJSON()
		if err != nil {
			return err
		}
		buf.Write(obj)

	}
	buf.WriteString(`,"hostname":`)
	fflib.WriteJsonString(buf, string(j.HostName))
	buf.WriteString(`,"appname":`)
	fflib.WriteJsonString(buf, string(j.AppName))
	buf.WriteString(`,"procid":`)
	fflib.WriteJsonString(buf, string(j.ProcId))
	buf.WriteString(`,"msgid":`)
	fflib.WriteJsonString(buf, string(j.MsgId))
	buf.WriteString(`,"structured":`)
	fflib.WriteJsonString(buf, string(j.Structured))
	buf.WriteString(`,"message":`)
	fflib.WriteJsonString(buf, string(j.Message))
	buf.WriteString(`,"properties":`)
	/* Falling back. type=map[string]map[string]string kind=map */
	err = buf.Encode(j.Properties)
	if err != nil {
		return err
	}
	buf.WriteByte('}')
	return nil
}

const (
	ffjtRegularSyslogbase = iota
	ffjtRegularSyslognosuchkey

	ffjtRegularSyslogFacility

	ffjtRegularSyslogSeverity

	ffjtRegularSyslogTimeReported

	ffjtRegularSyslogTimeGenerated

	ffjtRegularSyslogHostName

	ffjtRegularSyslogAppName

	ffjtRegularSyslogProcId

	ffjtRegularSyslogMsgId

	ffjtRegularSyslogStructured

	ffjtRegularSyslogMessage

	ffjtRegularSyslogProperties
)

var ffjKeyRegularSyslogFacility = []byte("facility")

var ffjKeyRegularSyslogSeverity = []byte("severity")

var ffjKeyRegularSyslogTimeReported = []byte("timereported")

var ffjKeyRegularSyslogTimeGenerated = []byte("timegenerated")

var ffjKeyRegularSyslogHostName = []byte("hostname")

var ffjKeyRegularSyslogAppName = []byte("appname")

var ffjKeyRegularSyslogProcId = []byte("procid")

var ffjKeyRegularSyslogMsgId = []byte("msgid")

var ffjKeyRegularSyslogStructured = []byte("structured")

var ffjKeyRegularSyslogMessage = []byte("message")

var ffjKeyRegularSyslogProperties = []byte("properties")

// UnmarshalJSON umarshall json - template of ffjson
func (j *RegularSyslog) UnmarshalJSON(input []byte) error {
	fs := fflib.NewFFLexer(input)
	return j.UnmarshalJSONFFLexer(fs, fflib.FFParse_map_start)
}

// UnmarshalJSONFFLexer fast json unmarshall - template ffjson
func (j *RegularSyslog) UnmarshalJSONFFLexer(fs *fflib.FFLexer, state fflib.FFParseState) error {
	var err error
	currentKey := ffjtRegularSyslogbase
	_ = currentKey
	tok := fflib.FFTok_init
	wantedTok := fflib.FFTok_init

mainparse:
	for {
		tok = fs.Scan()
		//	println(fmt.Sprintf("debug: tok: %v  state: %v", tok, state))
		if tok == fflib.FFTok_error {
			goto tokerror
		}

		switch state {

		case fflib.FFParse_map_start:
			if tok != fflib.FFTok_left_bracket {
				wantedTok = fflib.FFTok_left_bracket
				goto wrongtokenerror
			}
			state = fflib.FFParse_want_key
			continue

		case fflib.FFParse_after_value:
			if tok == fflib.FFTok_comma {
				state = fflib.FFParse_want_key
			} else if tok == fflib.FFTok_right_bracket {
				goto done
			} else {
				wantedTok = fflib.FFTok_comma
				goto wrongtokenerror
			}

		case fflib.FFParse_want_key:
			// json {} ended. goto exit. woo.
			if tok == fflib.FFTok_right_bracket {
				goto done
			}
			if tok != fflib.FFTok_string {
				wantedTok = fflib.FFTok_string
				goto wrongtokenerror
			}

			kn := fs.Output.Bytes()
			if len(kn) <= 0 {
				// "" case. hrm.
				currentKey = ffjtRegularSyslognosuchkey
				state = fflib.FFParse_want_colon
				goto mainparse
			} else {
				switch kn[0] {

				case 'a':

					if bytes.Equal(ffjKeyRegularSyslogAppName, kn) {
						currentKey = ffjtRegularSyslogAppName
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				case 'f':

					if bytes.Equal(ffjKeyRegularSyslogFacility, kn) {
						currentKey = ffjtRegularSyslogFacility
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				case 'h':

					if bytes.Equal(ffjKeyRegularSyslogHostName, kn) {
						currentKey = ffjtRegularSyslogHostName
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				case 'm':

					if bytes.Equal(ffjKeyRegularSyslogMsgId, kn) {
						currentKey = ffjtRegularSyslogMsgId
						state = fflib.FFParse_want_colon
						goto mainparse

					} else if bytes.Equal(ffjKeyRegularSyslogMessage, kn) {
						currentKey = ffjtRegularSyslogMessage
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				case 'p':

					if bytes.Equal(ffjKeyRegularSyslogProcId, kn) {
						currentKey = ffjtRegularSyslogProcId
						state = fflib.FFParse_want_colon
						goto mainparse

					} else if bytes.Equal(ffjKeyRegularSyslogProperties, kn) {
						currentKey = ffjtRegularSyslogProperties
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				case 's':

					if bytes.Equal(ffjKeyRegularSyslogSeverity, kn) {
						currentKey = ffjtRegularSyslogSeverity
						state = fflib.FFParse_want_colon
						goto mainparse

					} else if bytes.Equal(ffjKeyRegularSyslogStructured, kn) {
						currentKey = ffjtRegularSyslogStructured
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				case 't':

					if bytes.Equal(ffjKeyRegularSyslogTimeReported, kn) {
						currentKey = ffjtRegularSyslogTimeReported
						state = fflib.FFParse_want_colon
						goto mainparse

					} else if bytes.Equal(ffjKeyRegularSyslogTimeGenerated, kn) {
						currentKey = ffjtRegularSyslogTimeGenerated
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				}

				if fflib.EqualFoldRight(ffjKeyRegularSyslogProperties, kn) {
					currentKey = ffjtRegularSyslogProperties
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.EqualFoldRight(ffjKeyRegularSyslogMessage, kn) {
					currentKey = ffjtRegularSyslogMessage
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.EqualFoldRight(ffjKeyRegularSyslogStructured, kn) {
					currentKey = ffjtRegularSyslogStructured
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.EqualFoldRight(ffjKeyRegularSyslogMsgId, kn) {
					currentKey = ffjtRegularSyslogMsgId
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.SimpleLetterEqualFold(ffjKeyRegularSyslogProcId, kn) {
					currentKey = ffjtRegularSyslogProcId
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.SimpleLetterEqualFold(ffjKeyRegularSyslogAppName, kn) {
					currentKey = ffjtRegularSyslogAppName
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.EqualFoldRight(ffjKeyRegularSyslogHostName, kn) {
					currentKey = ffjtRegularSyslogHostName
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.SimpleLetterEqualFold(ffjKeyRegularSyslogTimeGenerated, kn) {
					currentKey = ffjtRegularSyslogTimeGenerated
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.SimpleLetterEqualFold(ffjKeyRegularSyslogTimeReported, kn) {
					currentKey = ffjtRegularSyslogTimeReported
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.EqualFoldRight(ffjKeyRegularSyslogSeverity, kn) {
					currentKey = ffjtRegularSyslogSeverity
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.SimpleLetterEqualFold(ffjKeyRegularSyslogFacility, kn) {
					currentKey = ffjtRegularSyslogFacility
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				currentKey = ffjtRegularSyslognosuchkey
				state = fflib.FFParse_want_colon
				goto mainparse
			}

		case fflib.FFParse_want_colon:
			if tok != fflib.FFTok_colon {
				wantedTok = fflib.FFTok_colon
				goto wrongtokenerror
			}
			state = fflib.FFParse_want_value
			continue
		case fflib.FFParse_want_value:

			if tok == fflib.FFTok_left_brace || tok == fflib.FFTok_left_bracket || tok == fflib.FFTok_integer || tok == fflib.FFTok_double || tok == fflib.FFTok_string || tok == fflib.FFTok_bool || tok == fflib.FFTok_null {
				switch currentKey {

				case ffjtRegularSyslogFacility:
					goto handle_Facility

				case ffjtRegularSyslogSeverity:
					goto handle_Severity

				case ffjtRegularSyslogTimeReported:
					goto handle_TimeReported

				case ffjtRegularSyslogTimeGenerated:
					goto handle_TimeGenerated

				case ffjtRegularSyslogHostName:
					goto handle_HostName

				case ffjtRegularSyslogAppName:
					goto handle_AppName

				case ffjtRegularSyslogProcId:
					goto handle_ProcId

				case ffjtRegularSyslogMsgId:
					goto handle_MsgId

				case ffjtRegularSyslogStructured:
					goto handle_Structured

				case ffjtRegularSyslogMessage:
					goto handle_Message

				case ffjtRegularSyslogProperties:
					goto handle_Properties

				case ffjtRegularSyslognosuchkey:
					err = fs.SkipField(tok)
					if err != nil {
						return fs.WrapErr(err)
					}
					state = fflib.FFParse_after_value
					goto mainparse
				}
			} else {
				goto wantedvalue
			}
		}
	}

handle_Facility:

	/* handler: j.Facility type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			j.Facility = string(string(outBuf))

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_Severity:

	/* handler: j.Severity type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			j.Severity = string(string(outBuf))

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_TimeReported:

	/* handler: j.TimeReported type=time.Time kind=struct quoted=false*/

	{
		if tok == fflib.FFTok_null {

			state = fflib.FFParse_after_value
			goto mainparse
		}

		tbuf, err := fs.CaptureField(tok)
		if err != nil {
			return fs.WrapErr(err)
		}

		err = j.TimeReported.UnmarshalJSON(tbuf)
		if err != nil {
			return fs.WrapErr(err)
		}
		state = fflib.FFParse_after_value
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_TimeGenerated:

	/* handler: j.TimeGenerated type=time.Time kind=struct quoted=false*/

	{
		if tok == fflib.FFTok_null {

			state = fflib.FFParse_after_value
			goto mainparse
		}

		tbuf, err := fs.CaptureField(tok)
		if err != nil {
			return fs.WrapErr(err)
		}

		err = j.TimeGenerated.UnmarshalJSON(tbuf)
		if err != nil {
			return fs.WrapErr(err)
		}
		state = fflib.FFParse_after_value
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_HostName:

	/* handler: j.HostName type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			j.HostName = string(string(outBuf))

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_AppName:

	/* handler: j.AppName type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			j.AppName = string(string(outBuf))

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_ProcId:

	/* handler: j.ProcId type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			j.ProcId = string(string(outBuf))

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_MsgId:

	/* handler: j.MsgId type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			j.MsgId = string(string(outBuf))

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_Structured:

	/* handler: j.Structured type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			j.Structured = string(string(outBuf))

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_Message:

	/* handler: j.Message type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			j.Message = string(string(outBuf))

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_Properties:

	/* handler: j.Properties type=map[string]map[string]string kind=map quoted=false*/

	{

		{
			if tok != fflib.FFTok_left_bracket && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for ", tok))
			}
		}

		if tok == fflib.FFTok_null {
			j.Properties = nil
		} else {

			j.Properties = make(map[string]map[string]string, 0)

			wantVal := true

			for {

				var k string

				var tmpJProperties map[string]string

				tok = fs.Scan()
				if tok == fflib.FFTok_error {
					goto tokerror
				}
				if tok == fflib.FFTok_right_bracket {
					break
				}

				if tok == fflib.FFTok_comma {
					if wantVal == true {
						// TODO(pquerna): this isn't an ideal error message, this handles
						// things like [,,,] as an array value.
						return fs.WrapErr(fmt.Errorf("wanted value token, but got token: %v", tok))
					}
					continue
				} else {
					wantVal = true
				}

				/* handler: k type=string kind=string quoted=false*/

				{

					{
						if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
							return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
						}
					}

					if tok == fflib.FFTok_null {

					} else {

						outBuf := fs.Output.Bytes()

						k = string(string(outBuf))

					}
				}

				// Expect ':' after key
				tok = fs.Scan()
				if tok != fflib.FFTok_colon {
					return fs.WrapErr(fmt.Errorf("wanted colon token, but got token: %v", tok))
				}

				tok = fs.Scan()
				/* handler: tmpJProperties type=map[string]string kind=map quoted=false*/

				{

					{
						if tok != fflib.FFTok_left_bracket && tok != fflib.FFTok_null {
							return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for ", tok))
						}
					}

					if tok == fflib.FFTok_null {
						tmpJProperties = nil
					} else {

						tmpJProperties = make(map[string]string, 0)

						wantVal := true

						for {

							var k string

							var tmpTmpJProperties string

							tok = fs.Scan()
							if tok == fflib.FFTok_error {
								goto tokerror
							}
							if tok == fflib.FFTok_right_bracket {
								break
							}

							if tok == fflib.FFTok_comma {
								if wantVal == true {
									// TODO(pquerna): this isn't an ideal error message, this handles
									// things like [,,,] as an array value.
									return fs.WrapErr(fmt.Errorf("wanted value token, but got token: %v", tok))
								}
								continue
							} else {
								wantVal = true
							}

							/* handler: k type=string kind=string quoted=false*/

							{

								{
									if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
										return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
									}
								}

								if tok == fflib.FFTok_null {

								} else {

									outBuf := fs.Output.Bytes()

									k = string(string(outBuf))

								}
							}

							// Expect ':' after key
							tok = fs.Scan()
							if tok != fflib.FFTok_colon {
								return fs.WrapErr(fmt.Errorf("wanted colon token, but got token: %v", tok))
							}

							tok = fs.Scan()
							/* handler: tmpTmpJProperties type=string kind=string quoted=false*/

							{

								{
									if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
										return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
									}
								}

								if tok == fflib.FFTok_null {

								} else {

									outBuf := fs.Output.Bytes()

									tmpTmpJProperties = string(string(outBuf))

								}
							}

							tmpJProperties[k] = tmpTmpJProperties

							wantVal = false
						}

					}
				}

				j.Properties[k] = tmpJProperties

				wantVal = false
			}

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

wantedvalue:
	return fs.WrapErr(fmt.Errorf("wanted value token, but got token: %v", tok))
wrongtokenerror:
	return fs.WrapErr(fmt.Errorf("ffjson: wanted token: %v, but got token: %v output=%s", wantedTok, tok, fs.Output.String()))
tokerror:
	if fs.BigError != nil {
		return fs.WrapErr(fs.BigError)
	}
	err = fs.Error.ToError()
	if err != nil {
		return fs.WrapErr(err)
	}
	panic("ffjson-generated: unreachable, please report bug.")
done:

	return nil
}
