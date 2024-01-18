package utils

import (
	"fmt"
	"strings"
)

type multiErrorList []error

// MultiErrorList returns an error which formats underlying errors as a list, or nil if err is nil.
func MultiErrorList(err error) (int, error) {
	if err == nil {
		return 0, nil
	}

	u, ok := err.(interface {
		Unwrap() []error
	})
	if !ok {
		return 1, multiErrorList{err}
	}
	errs := u.Unwrap()
	return len(errs), multiErrorList(errs)
}

func (m multiErrorList) Error() string {
	l := len(m)
	if l == 1 {
		return m[0].Error()
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "%d errors:", l)
	for _, e := range m {
		fmt.Fprintf(&sb, "\n\t- %v", e)
	}
	return sb.String()
}

func (m multiErrorList) Unwrap() []error {
	return m
}
