// +build !solution

package lab1

import (
	"strconv"
)

// Errors is an error returned by a multiwriter when there are errors with one
// or more of the writes. Errors will be in a one-to-one correspondence with
// the input elements; successful elements will have a nil entry.
//
// Should not be constructed if all entires are nil; in this case one should
// instead return just nil instead of a MultiError.
type Errors []error

/*
Task 4: Errors needed for multiwriter

You may find this blog post useful:
http://blog.golang.org/error-handling-and-go

Similar to a the Stringer interface, the error interface also defines a
method that returns a string.

type error interface {
    Error() string
}

Thus also the error type can describe itself as a string. The fmt package (and
many others) use this Error() method to print errors.

Implement the Error() method for the Errors type defined above.

For the following conditions should be covered.

1. When there are no errors in the slice, it should return:

"(0 errors)"

2. When there is one error in the slice, it should return:

The error string return by the corresponding Error() method.

3. When there are two errors in the slice, it should return:

The first error + " (and 1 other error)"

4. When there are X>1 errors in the slice, it should return:

The first error + " (and X other errors)"
*/
func (m Errors) Error() string {

	errList := Errors{} //Creates a list/slice that stores numbers of errors inputed

	for i := 0; i < len(m); i++ { //iterates all the "inputed" items
		switch m[i].(type) { //type assertions to check if the inputed values is an error
		case error: // If the inputed value is an error:
			errList = append(errList, m[i]) //add the value to the list/slice
		default: // If not an error:
			continue // Continue the loop/iteration
		}
	}
	errLen := len(errList) //To solve the test, check the len of the list and return given statements based on number of errors in list
	switch {
	case errLen == 0:
		return "(0 errors)"
	case errLen == 1:
		return errList[0].Error()
	case errLen == 2:
		return errList[0].Error() + " (and 1 other error)"
	case errLen > 2:
		return errList[0].Error() + " (and " + strconv.Itoa(errLen-1) + " other errors)"
	}
	return ""
}
