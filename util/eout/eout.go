package eout

import (
	"fmt"
	"log"
)

/* eout.Eout example invocations

if err != nil {
  return eout.Eout(err, "reading data")
}
if err = eout.Eout(err, "reading data"); err != nil {
  return nil, err
}
*/

type ErrorDescriber struct {
	environ string // the basic context that is potentially generating errors (like a GetThread function, the environ would be "get thread")
}

// parametrize Eout/Check such that error messages contain a defined context/environ
func Describe(environ string) ErrorDescriber {
	return ErrorDescriber{environ}
}

func (ed ErrorDescriber) Eout(err error, msg string, args ...interface{}) error {
	msg = fmt.Sprintf("%s: %s", ed.environ, msg)
	return Eout(err, msg, args...)
}

func (ed ErrorDescriber) Check(err error, msg string, args ...interface{}) {
	msg = fmt.Sprintf("%s: %s", ed.environ, msg)
	Check(err, msg, args...)
}

// format all errors consistently, and provide context for the error using the string `msg`
func Eout(err error, msg string, args ...interface{}) error {
	if err != nil {
		// received an invocation of e.g. format:
		// Eout(err, "reading data for %s and %s", "database item", "weird user")
		if len(args) > 0 {
			return fmt.Errorf("%s (%w)", fmt.Sprintf(msg, args...), err)
		}
		return fmt.Errorf("%s (%w)", msg, err)
	}
	return nil
}

func Check(err error, msg string, args ...interface{}) {
	if len(args) > 0 {
		err = Eout(err, msg, args...)
	} else {
		err = Eout(err, msg)
	}
	if err != nil {
		log.Fatalln(err)
	}
}
