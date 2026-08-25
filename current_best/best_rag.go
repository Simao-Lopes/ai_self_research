package main
import (
	"runtime"
)

func Search(docs [][]float32, query []float32, k int) []int {
	n := len(docs)
	if n == 0 || k <= 0 {
		return nil
	}
	if k > n {
		k = n
	}
	dim := len(query)
	cpus := runtime.GOMAXPROCS(0)
	if cpus < 1 {
		cpus = 1
	}

	nb := (n + cpus - 1) / cpus
	starts := make([]int, cpus+1)
	for c := 0; c <= cpus; c++ {
		s := c * nb
		if s > n {
			s = n
		}
		starts[c] = s
	}

	type res struct{ idx int32; sc float32 }
	perCpu := make([][]res, cpus)
	jobs := make(chan int, cpus)

	for c := 0; c < cpus; c++ {
		go func(c int) {
			qc := query
			for cid := range jobs {
				a := starts[cid]
				b := starts[cid+1]
				m := b - a
				if m == 0 {
					continue
				}
				L := make([]res, m)
				base := int32(a)
				for i := 0; i < m; i++ {
					di := docs[a+i]
					var s float32
					if len(di) == dim {
						for j := 0; j < dim; j++ {
							s += di[j] * qc[j]
						}
					} else if dl := len(di); dl > 0 {
						lm := dim
						if dl < lm {
							lm = dl
						}
						for j := 0; j < lm; j++ {
							s += di[j] * qc[j]
						}
					}
					L[i] = res{idx: base + int32(i), sc: s}
				}
				jobs <- cid // signal done (reuse channel)
				perCpu[cid] = L
			}
		}(c)
	}

	for c := 0; c < cpus; c++ {
		jobs <- c
	}
	for c := 0; c < cpus; c++ {
		<-jobs
	}
	close(jobs)

	out := make([]res, k)
	if k > 0 {
		var best float32 = -1 << 30
		first := true
		for cid := 0; cid < cpus; cid++ {
			L := perCpu[cid]
			tl := len(L)
			for i := 0; i < tl && (first || L[i].sc > best); i++ {
				cand := L[i]
				if first {
					out[0] = cand
					best = cand.sc
					first = false
					continue
				}
				pos := k - 1
				for pos >= 0 && out[pos].sc < cand.sc {
					if pos+1 < k {
						out[pos+1] = out[pos]
					}
					pos--
				}
				if pos+1 < k {
					out[pos+1] = cand
					best = out[k-1].sc
				}
			}
		}
	}

	resIdx := make([]int, len(out))
	for i, r := range out {
		resIdx[i] = int(r.idx)
	}
	return resIdx
}
