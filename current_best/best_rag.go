package main
import "sync"

const maxK = 256

type score struct {
	d float32
	i int
}

func Search(docs [][]float32, query []float32, k int) []int {
	n := len(docs)
	if n == 0 || k <= 0 {
		return nil
	}
	if k > maxK {
		k = maxK
	}
	q := make([]float64, len(query))
	for i, v := range query {
		q[i] = float64(v)
	}
	r := make([]score, n)
	var wg sync.WaitGroup
	chunk := (n + 15) / 16
	start := 0
	for s := 0; s < n; s += chunk {
		e := s + chunk
		if e > n {
			e = n
		}
		wg.Add(1)
		go func(a, b int) {
			defer wg.Done()
			base := a * len(query)
			for i := a; i < b; i++ {
				d := docs[i]
				if d == nil || len(d) != len(q) {
					continue
				}
				var acc float64 = 0
				p := base + (i-a)*len(query)
				for j, v := range q {
					acc += float64(d[j]) * v
				}
				r[i] = score{d: float32(acc), i: i}
				_ = p
			}
		}(start, e)
		start = e
	}
	wg.Wait()
	top := make([]score, k)
	copy(top, r[:k])
	siftDown(top, 0)
	for i := k; i < n; i++ {
		if r[i].d > top[0].d {
			top[0] = r[i]
			siftDown(top, 0)
		}
	}
	out := make([]int, k)
	for i, s := range top {
		out[k-1-i] = s.i
	}
	return out
}

func siftDown(h []score, i int) {
	n := len(h)
	v := h[i]
	for {
		l := 2*i + 1
		if l >= n {
			break
		}
		m := l
		r := l + 1
		if r < n && h[r].d < h[l].d {
			m = r
		}
		if v.d <= h[m].d {
			h[i] = v
			return
		}
		h[i] = h[m]
		i = m
	}
	h[i] = v
}
