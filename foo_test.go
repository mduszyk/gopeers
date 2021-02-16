package gopeers

import "testing"

func verifyFactorial(n int, expected int, t *testing.T) {
	got := factorial(n)
	if got != expected {
		t.Errorf("factorial(%d) = %d; want %d", n, got, expected)
	}
}

func TestFactorial(t *testing.T) {
	verifyFactorial(0, 1, t)
	verifyFactorial(1, 1, t)
	verifyFactorial(2, 2, t)
	verifyFactorial(3, 6, t)
	verifyFactorial(4, 24, t)
}
