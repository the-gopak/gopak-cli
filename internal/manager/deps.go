package manager

func topoOrder(nodes map[string][]string) ([]string, bool) {
	indeg := map[string]int{}
	out := map[string][]string{}
	for n, deps := range nodes {
		if _, ok := indeg[n]; !ok {
			indeg[n] = 0
		}
		for _, d := range deps {
			indeg[n] = indeg[n]
			indeg[d] = indeg[d]
			indeg[n]++
			out[d] = append(out[d], n)
		}
	}
	q := []string{}
	for n := range indeg {
		if indeg[n] == 0 {
			q = append(q, n)
		}
	}
	order := []string{}
	for len(q) > 0 {
		n := q[0]
		q = q[1:]
		order = append(order, n)
		for _, v := range out[n] {
			indeg[v]--
			if indeg[v] == 0 {
				q = append(q, v)
			}
		}
	}
	if len(order) != len(indeg) {
		return nil, false
	}
	return order, true
}
