// Copyright (c) 2022, Joshua J Baker. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package progsort

import (
	"math"
	"runtime"
	"sync"
	"sync/atomic"
)

// Sort sorts data given the provided less function.
//
// The nprocs param is number background goroutines used to help with
// sorting. Setting this to zero will use runtime.NumCPU().
//
// The spare param is a slice that the caller can provide for helping with the
// merge sort algorithm. When this is provided it must the same length as the
// data slice, also the caller is responsible to check the return value to
// determine if the data slice or the spare slice has the final sorted data.
// Setting spare to nil will have the sort operation manage the memory of the
// spare slice under-the-hood by allocating the needed memory, which also
// ensures that the data slice will always end up with the final sorted data.
//
// The prog param can be optionally provided if the caller want to monitor the
// continual progress of the sort operation. This param is a pointer to an
// int32 that can be then atomically read by the caller while the sort is in
// progress. The value of will be in the range of [0,math.MaxInt32].
// Reading the progress can be done like:
//
//     p := float64(atomic.LoadInt32(&prog))/math.MaxInt32
//
// Which will convert prog into a percentage between the range [0.0,1.0].
// Set prog to nil if progress monitoring is not needed.
//
// The cancel param can be used to cancel the sorting early. At any point
// while the sort is in progress, the cancel can be set to a non-zero value by
// the caller, and the sort will end early.
// This should be done atomically like:
//
//     atomic.StoreInt32(&cancel, 1)
//
// Set cancel to nil if cancellability is not needed.
func Sort[T any](
	data []T,
	less func(a, b T) bool,
	nprocs int,
	spare []T,
	prog *int32,
	cancel *int32,
) (swapped bool) {
	var spared bool
	if spare == nil {
		spare = make([]T, len(data))
	}
	if len(data) != len(spare) {
		panic("len(active) != len(spare)")
	}
	if nprocs <= 0 {
		nprocs = runtime.NumCPU()
	}
	var vproc int32
	if prog == nil {
		prog = &vproc
	}
	var vcancel int32
	if cancel == nil {
		cancel = &vcancel
	}
	swapped = mergeSort(data, spare, less, nprocs, prog, cancel)
	if swapped && spared {
		copy(data, spare)
		swapped = false
	}
	atomic.StoreInt32(prog, math.MaxInt32)
	return swapped
}

const pchunk = 1024

type mergeGroup struct {
	count  int
	i1, z1 int64
	i2, z2 int64
}

func addSteps(
	delta int64, prog *int32, cancel *int32,
	smu *sync.Mutex, steps *int64, nsteps int64,
) bool {
	if atomic.LoadInt32(cancel) != 0 {
		return false
	}
	smu.Lock()
	var vsteps int64
	*steps += delta
	vsteps = *steps
	perc := float64(vsteps) / float64(nsteps)
	atomic.StoreInt32(prog, int32(math.MaxInt32*perc))
	smu.Unlock()
	return true
}

func mergeSort[T any](
	active, spare []T,
	less func(a, b T) bool,
	nprocs int,
	prog *int32,
	cancel *int32,
) (swapped bool) {

	start, end := int64(0), int64(len(active))
	nmlevels := calcMergeLevels(end - start)
	nsteps := nmlevels * (end - start)
	var smu sync.Mutex
	var steps int64

	var datas [2][]T
	datas[0] = active
	datas[1] = spare

	var wg sync.WaitGroup
	var mergeC chan mergeGroup
	mergeC = make(chan mergeGroup, nprocs*16)
	defer func() {
		close(mergeC)
		wg.Wait()
	}()
	for g := 0; g < nprocs; g++ {
		go func() {
			var scounter int64
			for m := range mergeC {
				for i := 0; i < m.count; i++ {
					var ok bool
					scounter, ok = mergeSortUnit(
						m.i1, m.i1+m.z1, m.i2, m.i2+m.z2,
						active, spare, less, prog, cancel,
						&smu, &steps, nsteps,
						scounter,
					)
					if !ok {
						break
					}
					m.i1 += m.z1 + m.z2
					m.i2 = m.i1 + m.z1
				}
				if scounter > pchunk {
					if !addSteps(scounter, prog, cancel, &smu, &steps, nsteps) {
						break
					}
					scounter = 0
				}
				wg.Add(-m.count)
			}
		}()
	}

	var gm mergeGroup
	csize := int64(1)
	mlevel := int64(0)
	for ; mlevel < nmlevels; mlevel++ {
		active = datas[mlevel&1]
		spare = datas[(mlevel+1)&1]
		for i := start; i < end; {
			i1 := i
			i2 := i + csize
			size1 := csize
			size2 := csize
			if i2 > end {
				size1 = end - i1
				i2 = end
				size2 = 0
			} else if i2+size2 > end {
				size2 = end - i2
			}
			m := mergeGroup{i1: i1, z1: size1, i2: i2, z2: size2}
			if mlevel > 7 || size1 != csize || size2 != csize {
				wg.Add(1)
				m.count = 1
				mergeC <- m
			} else {
				if gm.count == 0 {
					gm = m
				}
				gm.count++
				if gm.count == 256>>mlevel {
					wg.Add(gm.count)
					mergeC <- gm
					gm.count = 0
				}
			}
			if atomic.LoadInt32(cancel) != 0 {
				break
			}
			i += size1 + size2
		}
		if gm.count > 0 {
			wg.Add(gm.count)
			mergeC <- gm
			gm.count = 0
		}
		wg.Wait()
		if atomic.LoadInt32(cancel) != 0 {
			break
		}
		csize *= 2
	}
	swapped = mlevel&1 == 1
	return swapped
}

func mergeSortUnit[T any](
	start1, end1 int64,
	start2, end2 int64,
	active, spare []T,
	less func(a, b T) bool,
	prog *int32,
	cancel *int32,
	smu *sync.Mutex,
	steps *int64,
	nsteps int64,
	scounter int64,
) (int64, bool) {
	const progFlush = 1024
	i := start1
	var a, b T
	var aset, bset bool
	for start1 < end1 && start2 < end2 {
		if !aset {
			a = active[start1]
		}
		if !bset {
			b = active[start2]
		}
		if less(b, a) {
			spare[i] = active[start2]
			start2++
			bset = false
		} else {
			spare[i] = active[start1]
			start1++
			aset = false
		}
		i++
		scounter++
		if scounter > pchunk {
			if !addSteps(scounter, prog, cancel, smu, steps, nsteps) {
				return 0, false
			}
			scounter = 0
		}
	}
	for start1 < end1 {
		spare[i] = active[start1]
		start1++
		i++
		scounter++
		if scounter > pchunk {
			if !addSteps(scounter, prog, cancel, smu, steps, nsteps) {
				return 0, false
			}
			scounter = 0
		}
	}
	for start2 < end2 {
		spare[i] = active[start2]
		start2++
		i++
		scounter++
		if scounter > pchunk {
			if !addSteps(scounter, prog, cancel, smu, steps, nsteps) {
				return 0, false
			}
			scounter = 0
		}
	}
	return scounter, true
}

func calcMergeLevels(count int64) int64 {
	// Calculate the number of levels needed to perform a merge sort.
	//
	// For example, let's say we have 22 inital items:
	//
	// 1: [.][.][.][.][.][.][.][.][.][.][.][.][.][.][.][.][.][.][.][.][.][.] 22
	// 2: [....][....][....][....][....][....][....][....][....][....][....] 11
	// 3: [..........][..........][..........][..........][..........][....]  6
	// 4: [......................][......................][................]  3
	// 5: [..............................................][................]  2
	// 6: [................................................................]  1
	//
	// This will take 5 levels to complete.
	//
	var levels int64
	for count > 1 {
		if count&1 == 0 {
			count /= 2
		} else {
			count = count/2 + 1
		}
		levels++
	}
	return levels
}
