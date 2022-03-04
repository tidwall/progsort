package progsort

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
)

func init() {
	seed, err := strconv.ParseInt(os.Getenv("SEED"), 10, 64)
	if err != nil {
		seed = time.Now().UnixNano()
	}
	fmt.Printf("SEED %d\n", seed)
	rand.Seed(seed)
}

func TestProgSort(t *testing.T) {
	start := time.Now()
	ilens := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	var sorted int
	var unsorted int
	for time.Since(start) < time.Second*2 {
		var N int
		var cancelEarly bool
		if len(ilens) > 0 {
			N = ilens[0]
			ilens = ilens[1:]
		} else {
			N = rand.Int() % 50_000
			cancelEarly = rand.Int()%5 == 0
		}
		items := rand.Perm(N)
		final := make([]int, N)
		var prog int32
		var cancel int32
		done := make(chan bool, 1)
		go func() {
			swapped := Sort(items, func(a, b int) bool {
				return a < b
			}, 0, final, &prog, &cancel)
			if swapped {
				items, final, swapped = final, items, !swapped
			}
			done <- true
		}()
		var prev float64
		for {
			p := float64(atomic.LoadInt32(&prog)) / math.MaxInt32
			if p < prev {
				t.Fatal("out of order")
			}
			if p > 0.5 && cancelEarly {
				atomic.StoreInt32(&cancel, 1)
				break
			}
			if p == 1 {
				break
			}
		}
		<-done
		if !sort.IntsAreSorted(items) {
			if !cancelEarly {
				t.Fatal("not sorted")
			} else {
				unsorted++
			}
		} else {
			sorted++
		}
	}
}

func BenchmarkInts(b *testing.B) {

	ilens := []int{
		100,
		500,
		1_000,
		5_000,
		10_000,
		50_000,
		100_000,
		500_000,
		1_000_000,
		5_000_000,
	}

	if os.Getenv("BIGCHART") != "" {
		ilens = append(ilens,
			10_000_000,
			50_000_000,
			100_000_000,
			500_000_000,
			1_000_000_000,
		)
	}

	b.Run("progsort", func(b *testing.B) {
		for _, n := range ilens {
			func(n int) {
				b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
					benchInts(b, n)
				})
			}(n)
		}
	})
	b.Run("stdlib", func(b *testing.B) {
		for _, n := range ilens {
			func(n int) {
				b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
					benchStdlibInts(b, n)
				})
			}(n)
		}
	})
}

func benchInts(b *testing.B, N int) {
	items := rand.Perm(N)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Sort(items, func(a, b int) bool { return a < b }, 0, nil, nil, nil)
	}
}

func benchStdlibInts(b *testing.B, N int) {
	items := rand.Perm(N)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sort.Ints(items)
	}
}
