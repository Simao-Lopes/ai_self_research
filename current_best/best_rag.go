package main
import "sync"

func Search(docs [][]float32, query []float32, k int) []int {
	n := len(docs)
	if n == 0 || k <= 0 {
		return nil
	}
	if k > n {
		k = n
	}
	dim := len(query)

	type d struct{ i int; v float32 }
	pool := make([]d, k)
	for j := range pool {
		pool[j] = d{i: -1, v: -1e38}
	}

	chunk := (n + 7) / 8
	workers := ((n-1)/chunk) + 1
	if workers > 8 {
		workers = 8
	}
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			start := w * chunk
			end := start + chunk
			if end > n {
				end = n
			}
			q := query
			for i := start; i < end; i++ {
				doc := docs[i]
				s := float32(0)
				m := dim & 7
				j := 0
				for ; j+8 <= m; j += 8 {
					s += q[j]*doc[j] + q[j+1]*doc[j+1] + q[j+2]*doc[j+2] + q[j+3]*doc[j+3] + q[j+4]*doc[j+4] + q[j+5]*doc[j+5] + q[j+6]*doc[j+6] + q[j+7]*doc[j+7]
				}
				for ; j < m; j++ {
					s += q[j] * doc[j]
				}
				if s > pool[k-1].v || (s == pool[k-1].v && i < pool[k-1].i) {
					pos := -1
					for p := k - 1; p >= 0; p-- {
						if pool[p].v <= s {
							pos = p
							break
						}
					}
					if pos == -1 {
						pool[k-1] = d{i: i, v: s}
					} else {
						for p := k - 1; p > pos; p-- {
							pool[p] = pool[p-1]
						}
						pool[pos] = d{i: i, v: s}
					}
				}
			}
		}(w)
	}
	wg.Wait()

	out := make([]int, k)
	for p := 0; p < k; p++ {
		if pool[p].i >= 0 {
			out[p] = pool[p].i
		} else {
			break
		}
	}
	return out[:k]
}
