package main
import "runtime"

func Search(docs [][]float32, query []float32, k int) []int {
	n := len(docs)
	d := len(query)
	if n == 0 || d == 0 || k <= 0 {
		return nil
	}
	if k > n {
		k = n
	}

	qnorm := 0.0
	for i := 0; i < d; i++ {
		v := float64(query[i])
		qnorm += v * v
	}
	qinv := 1.0 / (float64(qnorm) + 1e-20)

	if k == n {
		out := make([]int, n)
		for i := range out {
			out[i] = i
		}
		return out
	}

	scores := make([]float64, n)
	cpus := runtime.GOMAXPROCS(0)
	if cpus < 1 {
		cpus = 1
	}
	if cpus > 8 {
		cpus = 8
	}
	chunk := (n + cpus - 1) / cpus
	for s := 0; s < cpus; s++ {
		start := s * chunk
		end := start + chunk
		if end > n {
			end = n
		}
		if start >= end {
			break
		}
		go func(lo, hi int) {
			for i := lo; i < hi; i++ {
				vec := docs[i]
				dn := 0.0
				for j := 0; j < d; j++ {
					v := float64(vec[j])
					dn += v * v
				}
				dot := 0.0
				for j := 0; j < d; j++ {
					dot += float64(vec[j]) * float64(query[j])
				}
				scores[i] = dot / (float64(dn)*qinv) // dot / (dnorm*qnrm)
			}
		}(start, end)
	}

	res := make([]int, k)
	minv := -1e308
	for i := 0; i < k; i++ {
		if scores[i] > minv {
			minv = scores[i]
			res[k-1-i] = i // will fix order below
		}
	}

	idxs := make([]int, n)
	vls := make([]float64, n)
	for i := range idxs {
		idxs[i] = i
		vls[i] = scores[i]
	}

	pk := k
	if pk > len(vls) {
		pk = len(vls)
	}

	// heap select top k using insertion into fixed array (partial sort of first k, then maintain min-heap)
	topIdx := make([]int, pk)
	topVal := make([]float64, pk)
	for i := 0; i < pk; i++ {
		topIdx[i] = idxs[i]
		topVal[i] = vls[i]
	}

	// build min-heap on topVal
	pk2 := len(topVal)
	for i := pk2/2 - 1; i >= 0; i-- {
		siftDownMin(topVal, topIdx, i, pk2)
	}

	for i := pk; i < n; i++ {
		if vls[i] > topVal[0] {
			topVal[0] = vls[i]
			topIdx[0] = idxs[i]
			siftDownMin(topVal, topIdx, 0, len(topVal))
		}
	}

	out := make([]int, k)
	for i := range out {
		out[i] = topIdx[i]
	}
	return out
}

func siftDownMin(vals []float64, idxs []int, i, n int) {
	for {
		left := 2*i + 1
		if left >= n {
			break
		}
		right := left + 1
		small := left
		if right < n && vals[right] < vals[left] {
			small = right
		}
		if vals[i] <= vals[small] {
			break
		}
		vals[i], vals[small] = vals[small], vals[i]
		idxs[i], idxs[small] = idxs[small], idxs[i]
		i = small
	}
}
