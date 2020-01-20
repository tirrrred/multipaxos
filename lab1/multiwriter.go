// +build !solution

package lab1

import (
	"io"
)

/*
Task 5: WriteTo function for multiple writers

In this task you are going to implement a WriteTo function that writes to
multiple writers. This is similar to the io.MultiWriter() function. However,
the io.MultiWriter() function can only return a single error, and it is not
possible to figure out which of the original writers caused the error. That
is, you must use the Errors type that you developed for Task 4.

Implement the WriteTo() function defined below.

For the following conditions should be satisfied.

1. Write the []byte slice to all writers.

2. The function should return (using n) the bytes written by each writer with
index position corresponding to the index position of the writer. An empty
slice ([]int{}) should be returned if there are no writers as argument to the
function.

3. If one of the writers returned an error, that error should be returned in
the index position corresponding to the index position of the writer.

4. If one of the writers could not write the entire buffer, the error
io.ErrShortWrite should be returned in the index position corresponding to
that writer's index position.

5. If no errors were observed, the function must return n, nil.
*/

// WriteTo writes b to the provided writers, returns a slice of the number
// of byte written to each writer, and a slice of errors, if any.

//type Errors []error

func WriteTo(b []byte, writers ...io.Writer) (n []int, errs Errors) {

	//condition 2
	if len(writers) == 0 {
		return []int{}, nil
	}

	//condition 1
	for _, writer := range writers { //Iterates over the content in the slice writers
		numb, e := writer.Write(b) //Writes b []byte to each writer in the slice writers.

		//condition 3
		n = append(n, numb) //Adds the returned n from writer.Write(b) to the n slice
		//condition 4
		if numb < len(b) {
			errs = append(errs, io.ErrShortWrite)
			continue
		}
		errs = append(errs, e) //Adds the returned errs from writer.Write(b) to the errs slice
	}
	//condition 5
	if errs.Error() == "(0 errors)" {
		return n, nil
	}

	return n, errs
}
