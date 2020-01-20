//+build !solution

package lab1

// Task 1: Fibonacci numbers
//
// fibonacci(n) returns the n-th Fibonacci number, and is defined by the
// recurrence relation F_n = F_n-1 + F_n-2, with seed values F_0=0 and F_1=1.
func fibonacci(n uint) uint {
	var F0 uint = 0 //seed one
	var F1 uint = 1 //seed two

	//create a slice consisting of fibonacci numbers up to the n-th fib-number (given as an input)
	fibNumbers := []uint{F0, F1}

	//for loops that appends the fibonacci numbers based on the seed numbers
	for i := uint(0); i <= n; i++ {
		fibNumbers = append(fibNumbers, fibNumbers[i]+fibNumbers[i+1])
	}
	//returns the n-th Fibonacci number
	return fibNumbers[n]
}