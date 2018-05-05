package eerrors

import (
	"fmt"
	"io"
	"strings"
	"sync/atomic"
	"unsafe"

	errors "github.com/segmentio/errors-go"
)

type errNode struct {
	err  error
	next *errNode
}

func writePrefixLine(w io.Writer, prefix string, s string) {
	first := true
	for len(s) > 0 {
		if first {
			first = false
		} else {
			io.WriteString(w, prefix)
		}

		idx := strings.IndexByte(s, '\n')
		if idx < 0 {
			idx = len(s) - 1
		}

		io.WriteString(w, s[:idx+1])
		s = s[idx+1:]
	}
}

type ErrorSlice []error

func (s ErrorSlice) Empty() bool {
	return len(s) == 0
}

func (s ErrorSlice) Error() string {
	if len(s) == 0 {
		return ""
	}
	var buf strings.Builder
	first := true
	for _, item := range s {
		if first {
			first = false
		} else {
			buf.WriteString("; ")
		}
		buf.WriteString(item.Error())
	}
	return buf.String()
}

func (s ErrorSlice) Format(f fmt.State, c rune) {
	if len(s) == 0 {
		return
	}
	if c == 'v' && f.Flag('+') {
		io.WriteString(f, "the following errors occurred:")
		for _, item := range s {
			io.WriteString(f, "\n -  ")
			writePrefixLine(f, "    ", fmt.Sprintf("%+v", item))
		}
	} else {
		io.WriteString(f, s.Error())
	}
}

func (s ErrorSlice) Causes() []error {
	if len(s) == 0 {
		return nil
	}
	s2 := make([]error, 0, len(s))
	for _, err := range s {
		s2 = append(s2, err)
	}
	return s2
}

func (s ErrorSlice) WithTags(tags ...string) error {
	if len(s) == 0 {
		return nil
	}
	if len(tags) < 2 {
		return s
	}
	keyFlag := true
	key := ""
	value := ""
	etags := make([]errors.Tag, 0, len(tags)/2)
	for _, tag := range tags {
		if keyFlag {
			key = tag
		} else {
			value = tag
			etags = append(etags, errors.T(key, value))
		}
		keyFlag = !keyFlag
	}
	if len(tags) == 0 {
		return s
	}
	return errors.WithTags(s, etags...)
}

func (s ErrorSlice) Wrap(msg string) error {
	if len(s) == 0 {
		return nil
	}
	return errors.Wrap(s, msg)
}

func (s ErrorSlice) WithTypes(types ...string) error {
	if len(s) == 0 {
		return nil
	}
	return errors.WithTypes(s, types...)
}

func (s ErrorSlice) With(msg string, types []string, tags ...string) (err error) {
	if len(s) == 0 {
		return nil
	}
	err = s.WithTags(tags...)
	err = errors.WithTypes(err, types...)
	return errors.Wrap(err, msg)
}

func (s ErrorSlice) Is(typ string) bool {
	if len(s) == 0 {
		return false
	}
	for _, cause := range s {
		if errors.Is(typ, cause) {
			return true
		}
	}
	return false
}

type ChainedErrors struct {
	head *errNode
	tail *errNode
}

func ChainErrors() (c *ChainedErrors) {
	stub := new(errNode)
	c = &ChainedErrors{
		head: stub,
		tail: stub,
	}
	return c
}

func (c *ChainedErrors) Append(s ...error) *ChainedErrors {
	for _, e := range s {
		c.append(e)
	}
	return c
}

func (c *ChainedErrors) append(err error) *ChainedErrors {
	if err == nil {
		return c
	}
	if e, ok := err.(causes); ok {
		if e == nil {
			return c
		}
		return c.Extend(e.Causes())
	}

	n := new(errNode)
	n.err = err
	(*errNode)(atomic.SwapPointer((*unsafe.Pointer)(unsafe.Pointer(&c.head)), unsafe.Pointer(n))).next = n
	return c
}

func (c *ChainedErrors) get() error {
	tail := c.tail
	next := tail.next
	if next != nil {
		(*errNode)(atomic.SwapPointer((*unsafe.Pointer)(unsafe.Pointer(&c.tail)), unsafe.Pointer(next))).err = next.err
		return next.err
	}
	return nil
}

func (c *ChainedErrors) Sum() (s ErrorSlice) {
	s = make([]error, 0, 10)
	var err error
	for {
		err = c.get()
		if err == nil {
			break
		}
		s = append(s, err)
	}
	if len(s) == 0 {
		return nil
	}
	return s
}

func (c *ChainedErrors) Extend(s []error) *ChainedErrors {
	for _, e := range s {
		c.append(e)
	}
	return c
}

func (c *ChainedErrors) Receive(ch <-chan error) *ChainedErrors {
	for e := range ch {
		c.append(e)
	}
	return c
}

func (c *ChainedErrors) Send(ch chan<- error) {
	var err error
	for {
		err = c.get()
		if err == nil {
			break
		}
		ch <- err
	}
}

func Combine(errs ...error) (err error) {
	return ChainErrors().Extend(errs).Sum()
}