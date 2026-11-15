// Проверка циклов после объединения путей, найденных на разной глубине.
package graph

const (
	visiting = 1
	finished = 2
)

// detectCycles проверяет итоговые рёбра: более короткий путь мог замкнуть новый цикл.
func (g *Graph) detectCycles() {
	states := make(map[string]int)
	for _, id := range g.Order {
		if states[id] == 0 {
			g.walkCycles(id, states)
		}
	}
}

// walkCycles рекурсивно отмечает узлы текущего стека и сохраняет обратные рёбра.
func (g *Graph) walkCycles(id string, states map[string]int) {
	states[id] = visiting
	for _, child := range g.Edges[id] {
		switch states[child] {
		case visiting:
			g.Cycles = append(g.Cycles, Edge{From: id, To: child})
		case 0:
			g.walkCycles(child, states)
		}
	}
	states[id] = finished
}
