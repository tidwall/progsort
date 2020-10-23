// Copyright (c) 2020, Joshua J Baker. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package progsort

import (
	"math"
	"reflect"
	"sort"
)

// Slice sorts a slice and provides approximate progress.
// This operation works the same as sort.Slice.
// The sorted param is an estimate of how presorted the data is, and the value
// should be between 0.0 and 1.0 where 0.0 being completely random and 1.0
// being totally sorted. It's just an estimate and it's fine to pass 0.0.
// The prog function returns the progress at various steps during the sort
// operation. It's always increasing, but because it's approximate, the final
// value will likely not be exactly 1.0.
func Slice(slice interface{}, less func(i, j int) bool, prog func(f float64), sorted float64) {
	len := float64(reflect.ValueOf(slice).Len())
	if len == 0 {
		return
	}
	sorted = math.Max(math.Min(sorted, 1.0), 0.0)
	x := math.Log(len) * (((1.5 - 1.28403) * (1 - sorted)) + 1.28403)
	var step int
	sort.Slice(slice, func(i, j int) bool {
		step++
		if len < 4096 || step&255 == 0 {
			prog(float64(step) / len / x)
		}
		return less(i, j)
	})
	if len >= 4096 && step&255 != 0 {
		prog(float64(step) / len / x)
	}
}
