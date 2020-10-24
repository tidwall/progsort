# progsort

[![GoDoc](https://img.shields.io/badge/api-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/tidwall/progsort)

Sorts a slice and provides approximate and continual progress while the
operation is running.

## Install

```
go get -u github.com/tidwall/progsort
```

## Example

```go
package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/tidwall/progsort"
)

func main() {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	nums := make([]float64, 1000000)
	for i := range nums {
		nums[i] = rng.Float64()
	}
	progsort.Slice(nums, func(i, j int) bool {
		return nums[i] < nums[j]
	}, func(prog float64) {
		fmt.Printf("\rSorting... %0.1f%% ", prog*100)
	}, 0)
	fmt.Println()
}

```
