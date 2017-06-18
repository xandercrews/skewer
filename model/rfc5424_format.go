package model

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

func ParseRfc5424Format(m string, dont_parse_sd bool) (*SyslogMessage, error) {
	// HEADER = PRI VERSION SP TIMESTAMP SP HOSTNAME SP APP-NAME SP PROCID SP MSGID
	// PRI = "<" PRIVAL ">"
	// SYSLOG-MSG = HEADER SP STRUCTURED-DATA [SP MSG]

	smsg := SyslogMessage{}
	splits := strings.SplitN(m, " ", 7)

	if len(splits) < 7 {
		return nil, fmt.Errorf("Message does not have enough parts")
	}

	var err error
	smsg.Priority, smsg.Facility, smsg.Severity, smsg.Version, err = ParsePriorityVersion(splits[0])
	if err != nil {
		return nil, err
	}

	n := time.Now()
	if splits[1] == "-" {
		smsg.TimeReported = time.Now()
	} else {
		t1, err := time.Parse(time.RFC3339Nano, splits[1])
		if err != nil {
			t2, err := time.Parse(time.RFC3339, splits[1])
			if err != nil {
				smsg.TimeReported = n
			} else {
				smsg.TimeReported = t2
			}
		} else {
			smsg.TimeReported = t1
		}
	}
	smsg.TimeGenerated = n

	if splits[2] != "-" {
		smsg.Hostname = splits[2]
	}
	if splits[3] != "-" {
		smsg.Appname = splits[3]
	}
	if splits[4] != "-" {
		smsg.Procid = splits[4]
	}
	if splits[5] != "-" {
		smsg.Msgid = splits[5]
	}
	structured_and_msg := strings.TrimSpace(splits[6])
	if strings.HasPrefix(structured_and_msg, "-") {
		// structured data is empty
		smsg.Message = strings.TrimSpace(structured_and_msg[1:])
	} else if strings.HasPrefix(structured_and_msg, "[") {
		s1, s2, err := SplitStructuredAndMessage(structured_and_msg)
		if err != nil {
			return nil, err
		}
		smsg.Message = s2
		smsg.Properties = map[string]interface{}{}
		if dont_parse_sd {
			smsg.Structured = s1
		} else {
			smsg.Structured = ""
			props, err := ParseStructuredData(s1)
			if err != nil {
				return nil, err
			}
			if props != nil {
				smsg.Properties["rfc5424-sd"] = props
			}
		}
	} else {
		return nil, fmt.Errorf("Invalid structured data")
	}

	return &smsg, nil
}

func SplitStructuredAndMessage(structured_and_msg string) (string, string, error) {
	length := len(structured_and_msg)
	for i := 0; i < length; i++ {
		if structured_and_msg[i] == ']' {
			if i == (length - 1) {
				return structured_and_msg, "", nil
			}
			if structured_and_msg[i+1] == ' ' {
				return structured_and_msg[:i+1], strings.TrimSpace(structured_and_msg[i+1:]), nil
			}
		}
	}
	return "", "", fmt.Errorf("Invalid structured data")
}

func ParsePriorityVersion(pv string) (Priority, Facility, Severity, Version, error) {
	if pv[0] != byte('<') {
		return 0, 0, 0, 0, fmt.Errorf("Invalid priority")
	}
	i := strings.Index(pv, ">")
	if i < 2 {
		return 0, 0, 0, 0, fmt.Errorf("Invalid priority")
	}
	if len(pv) <= (i + 1) {
		return 0, 0, 0, 0, fmt.Errorf("Invalid priority")
	}
	p, err := strconv.Atoi(pv[1:i])
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("Invalid priority")
	}
	f := Facility(p / 8)
	s := Severity(p % 8)
	v, err := strconv.Atoi(pv[i+1:])
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("Invalid priority")
	}

	return Priority(p), f, s, Version(v), nil
}

func ParseStructuredData(sd string) (m map[string]map[string]string, err error) {
	// see https://tools.ietf.org/html/rfc5424#section-6.3
	if !utf8.ValidString(sd) {
		return nil, fmt.Errorf("Structured data is not utf-8 encoded")
	}
	m = map[string]map[string]string{}
	l := len(sd)
	position := 0
	current_sdid := ""
	current_name := ""

	var openBracket func() error
	var sdid func() error
	var value func() error
	var param func() error

	value = func() error {
		// a bit long and painful to take care of escaped characters
		if position == l {
			return fmt.Errorf("Expected SD-VALUE, got nothing")
		}
		if sd[position] != byte('"') {
			return fmt.Errorf("SD-VALUE should start with a quote")
		}
		position++
		p := position
		found := false
		for p < l && !found {
			if sd[p] == byte('\\') {
				p++
				if p >= l {
					return fmt.Errorf("Unexpected end after a \\")
				}
				if sd[p] == byte('"') || sd[p] == byte('\\') || sd[p] == byte(']') {
					p++
				}
			} else if sd[p] == byte('"') {
				found = true
			} else {
				p++
			}
		}
		if found {
			val := sd[position:p]
			m[current_sdid][current_name] = val
			fmt.Printf("%s//%s//%s\n", current_sdid, current_name, val)
			position += len(val)
			position++ // count for the closing quote
			if position >= l {
				return fmt.Errorf("Abrupt end of SD-ELEMENT")
			}
			if sd[position] == byte(' ') {
				position++
				return param()
			} else if sd[position] == byte(']') {
				position++
				return openBracket()
			} else {
				return fmt.Errorf("Expected SP or ']' but got '%s' instead", string(sd[position]))
			}

		} else {
			return fmt.Errorf("The end of SD-VALUE was not found")
		}
	}

	param = func() error {
		name_end := strings.Index(sd[position:], "=")
		if name_end < 1 {
			return fmt.Errorf("Invalid SD-NAME")
		}
		current_name = sd[position : position+name_end]
		position += name_end
		position++ // count the =
		return value()
	}

	sdid = func() error {
		end := strings.IndexAny(sd[position:], " ]")
		if end < 1 {
			return fmt.Errorf("Invalid SDID")
		}
		current_sdid = sd[position : position+end]
		position += end
		m[current_sdid] = map[string]string{}
		if sd[position] == byte(' ') {
			// now read the params
			position++
			return param()
		} else {
			// end of the element
			position++
			return openBracket()
		}

	}

	openBracket = func() error {
		if position == l {
			return nil
		}
		if sd[position] == byte('[') {
			position++
			return sdid()
		} else {
			return fmt.Errorf("Expected '[' but got '%s' instead", string(sd[position]))
		}
	}

	err = openBracket()
	if err != nil {
		return nil, err
	}

	return m, nil
}
