package dag

func (g *Graph) CriticalPath() []string {
	order, err := g.TopologicalSort()
	if err != nil {
		return nil
	}

	dist := make(map[string]int, len(order))
	prev := make(map[string]string, len(order))

	for _, id := range order {
		dist[id] = 1
		prev[id] = ""
	}

	for _, u := range order {
		for _, v := range g.Adjacency[u] {
			if dist[u]+1 > dist[v] {
				dist[v] = dist[u] + 1
				prev[v] = u
			}
		}
	}

	var endNode string
	maxDist := 0
	for _, id := range order {
		if dist[id] > maxDist {
			maxDist = dist[id]
			endNode = id
		}
	}

	if endNode == "" {
		return nil
	}

	var path []string
	for node := endNode; node != ""; node = prev[node] {
		path = append(path, node)
	}

	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	return path
}
