package main
import "sync"

func Search(docs [][]float32, query []float32, k int) []int {
	n := len(docs)
	if n == 0 || k <= 0 {
		return nil
	}
	dim := len(query)
	if dim == 0 {
		return make([]int, k)
	}
	if k > n {
		k = n
	}

	maxWorkers := 8
	nw := maxWorkers * 2
	if nw < 1 {
		nw = 1
	}
	if nw > n {
		nw = n
	}
	cs := (n + nw - 1) / nw

	type pr struct {
		v float32
		i int32
	}

	out := make([]int, k)
	scores := make([]float32, k)
	idxs := make([]int32, k)
	for i := range out {
		out[i] = -1
	}

	var wg sync.WaitGroup
	for w := 0; w < nw; w++ {
		lo := w * cs
		if lo >= n {
			break
		}
		hi := lo + cs
		if hi > n {
			hi = n
		}
		wg.Add(1)
		go func(lo, hi int) {
			defer wg.Done()

			localScores := make([]float32, k)
			localIdxs := make([]int32, k)
			for i := range localScores {
				localScores[i] = -1e30
				localIdxs[i] = -1
			}

			for r := lo; r < hi; r++ {
				doc := docs[r]
				if len(doc) != dim {
					continue
				}
				s := float32(0)
				_ = doc[:dim]
				for j := 0; j < dim; j += 8 {
					e := dim - j
					var a0, a1, a2, a3, a4, a5, a6, a7 float32
					if e >= 8 {
						a0 = doc[j+0] * query[j+0]
						a1 = doc[j+1] * query[j+1]
						a2 = doc[j+2] * query[j+2]
						a3 = doc[j+3] * query[j+3]
						a4 = doc[j+4] * query[j+4]
						a5 = doc[j+5] * query[j+5]
						a6 = doc[j+6] * query[j+6]
						a7 = doc[j+7] * query[j+7]
					} else {
						for m := 0; m < e; m++ {
							switch m {
							case 0:
								a0 = doc[j+m] * query[j+m]
							case 1:
								a1 = doc[j+m] * query[j+m]
							case 2:
								a2 = doc[j+m] * query[j+m]
							case 3:
								a3 = doc[j+m] * query[j+m]
							case 4:
								a4 = doc[j+m] * query[j+m]
							case 5:
								a5 = doc[j+m] * query[j+m]
							case 6:
								a6 = doc[j+m] * query[j+m]
							}
						}
					}
					s += (a0 + a1) + (a2 + a3) + (a4 + a5) + (a6 + a7)
				}

				pos := k - 1
				for pos > 0 && s > localScores[pos-1] {
					localScores[pos], localIdxs[pos] = localScores[pos-1], localIdxs[pos-1]
					pos--
				}
				if s >= localScores[0] || k == n {
					localScores[pos] = s
					localIdxs[pos] = int32(r)
				}
			}

			for i := 0; i < k; i++ {
				s := localScores[i]
				if s <= -1e29 && localIdxs[i] == -1 {
					continue
				}
				pos := k - 1
				for pos > 0 && s > scores[pos-1] {
					scores[pos], idxs[pos] = scores[pos-1], idxs[pos-1]
					pos--
				}
				if s >= scores[0] || k == n {
					scores[pos] = s
					idxs[pos] = localIdxs[i]
				}
			}
		}(lo, hi)
	}
	wg.Wait()

	for i := 0; i < k; i++ {
		if idxs[i] >= 0 && idxs[i] < int32(n) {
			out[i] = int(idxs[i])
		} else {
			break
		}
	}
	return out[:k]
}
