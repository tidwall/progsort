package progsort

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestSlice(t *testing.T) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	const N = 1000000
	nums := make([]float64, N)
	for i := 0; i < N; i++ {
		nums[i] = rng.Float64()
	}
	var last float64
	Slice(nums[:N], func(i, j int) bool {
		return nums[i] < nums[j]
	}, func(prog float64) {
		if prog < last {
			t.Fatal("out of order")
		}
	}, 0)

	{
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		const N = 1000000
		nums := make([]float64, N)
		for i := 0; i < N; i++ {
			nums[i] = rng.Float64()
		}
		Slice(nums[:N], func(i, j int) bool {
			return nums[i] < nums[j]
		}, func(prog float64) {
			fmt.Printf("\rSorting... %0.1f%% ", prog*100)
		}, 0)
		fmt.Println()
	}
}
