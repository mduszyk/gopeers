package peers

import (
	"fmt"
	"math"
	"testing"
)

func TestTreeFind(t *testing.T) {
	a := make([]int, 20)
	b := []int{1, 2, 3}
	c := []int{4, 5, 6, 7}
	copy(a, b)
	copy(a[len(b):], c)
	fmt.Printf("%d, %v\n", len(a), a[:int(math.Min(float64(len(a)), 30))])
}
