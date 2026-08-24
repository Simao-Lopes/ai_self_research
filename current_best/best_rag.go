package main
import "sync"

func dot32(a, b []float32) float32 {
	var s float32
	for i := range a {
		s += a[i] * b[i]
	}
	return s
}

func Search(docs [][]float32, query []float32, k int) []int {
	n := len(docs)
	if n == 0 || k <= 0 {
		return nil
	}
	if k > n {
		k = n
	}

	numGoroutines := 4
	chunkSize := (n + numGoroutines - 1) / numGoroutines

	type partial struct{ idx int }
	var wg sync.WaitGroup
	results := make([][]int, numGoroutines)
	scores := make([][]float32, numGoroutines)

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			start := g * chunkSize
			end := start + chunkSize
			if end > n {
				end = n
			}
			localIdx := make([]int, 0, end-start)
			localScores := make([]float32, 0, end-start)
			for i := start; i < end; i++ {
				d := docs[i]
				if len(d) == len(query) {
					var s float32
					for j := range query {
						s += d[j] * query[j]
					}
					localScores = append(localScores, s)
					localIdx = append(localIdx, i)
				}
			}
			results[g] = localIdx
			scores[g] = localScores
		}(g)
	}
	wg.Wait()

	var finalIdx []int
	var finalScores []float32
	for g := 0; g < numGoroutines; g++ {
		if len(results[g]) > 0 {
			finalIdx = append(finalIdx, results[g]...)
			finalScores = append(finalScores, scores[g]...)
		}
	}

	m := len(finalIdx)
	if m <= k {
		out := make([]int, m)
		copy(out, finalIdx)
		return out
	}

	type pair struct {
		s   float32
		idx int
	}
	hp := make([]pair, 0, k+1)
	push := func(s float32, i int) {
		if len(hp) < k {
			hp = append(hp, pair{s, i})
			return
		}
		if s > hp[0].s {
			hp[0] = pair{s, i}
			pos := 1
			for pos < len(hp) {
				lc, rc := pos*2+1, pos*2+2
				midx := pos
				if lc < len(hp) && hp[lc].s > hp[midx].s {
					midx = lc
				}
				if rc < len(hp) && hp[rc].s > hp[midx].s {
					midx = rc
				}
				if midx == pos {
					break
				}
				hp[pos], hp[midx] = hp[midx], hp[pos]
				pos = midx
			}
		}
	}

	for i := 0; i < k; i++ {
		push(finalScores[i], finalIdx[i])
	}
	if len(hp) == k {
		i := (len(hp) - 1) / 2
		for i >= 0 {
			lc, rc := i*2+1, i*2+2
			midx := i
			if lc < len(hp) && hp[lc].s > hp[midx].s {
				midx = lc
			}
			if rc < len(hp) && hp[rc].s > hp[midx].s {
				midx = rc
			}
			if midx == i {
				break
			}
			hp[i], hp[midx] = hp[midx], hp[i]
			i = midx
		}
		for i := k; i < m; i++ {
			push(finalScores[i], finalIdx[i])
		}
	}

	out := make([]int, 0, len(hp))
	tmp := make([]pair, len(hp))
	copy(tmp, hp)
	for len(tmp) > 0 {
		top := tmp[0]
		out = append(out, top.idx)
		last := tmp[len(tmp)-1]
		tmp = tmp[:len(tmp)-1]
		if len(tmp) == 0 {
			break
		}
		tmp[0] = last
		pos := 1
		for pos < len(tmp) {
			lc, rc := pos*2+1, pos*2+2
			midx := pos
			if lc < len(tmp) && tmp[lc].s > tmp[midx].s {
				midx = lc
			}
			if rc < len(tmp) && tmp[rc].s > tmp[midx].s {
				midx = rc
			}
			if midx == pos {
				break
			}
			tmp[pos], tmp[midx] = tmp[midx], tmp[pos]
			pos = midx
		}
	}

	for i := 0; i < len(out)/2; i++ {
		out[i], out[len(out)-1-i] = out[len(out)-1-i], out[i]
	}
	return out
}
