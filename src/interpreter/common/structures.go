package common

type BinaryTree[T any] struct {
	Left  *BinaryTree[T]
	Right *BinaryTree[T]
	Value T
}

/*
 * Perform a depth-first search on the binary tree, calling a given callback function for each node.
 * If the function returns `true`, the search will stop and that node and `true` will be returned.
 * If the function always returns `false`, `nil` and `false` will be returned.
 */
func (tree *BinaryTree[T]) DFS(finder func(node *BinaryTree[T]) bool) (*BinaryTree[T], bool) {
	stack := []*BinaryTree[T]{tree}

	for len(stack) > 0 {
		node := stack[len(stack)-1]

		stack = stack[:len(stack)-1]

		// finder could modify node.left and/or node.right, so we get them before we call it
		left, right := node.Left, node.Right

		if finder(node) {
			return node, true
		}

		if right != nil {
			stack = append(stack, right)
		}

		if left != nil {
			stack = append(stack, left)
		}
	}

	return nil, false
}

func NewBalancedBinaryTreeFromSlice[T any](slice []T) *BinaryTree[T] {
	if len(slice) == 0 {
		return nil
	}

	if len(slice) == 1 {
		return &BinaryTree[T]{
			Value: slice[0],
		}
	}

	middle := len(slice) / 2

	return &BinaryTree[T]{
		Left:  NewBalancedBinaryTreeFromSlice(slice[:middle]),
		Right: NewBalancedBinaryTreeFromSlice(slice[middle:]),
	}
}

type Graph[T any] struct {
	nodes []T
	edges map[int][]int
}

func (graph *Graph[T]) AddEdge(i int, j int) {
	if edges, ok := graph.edges[i]; ok {
		graph.edges[i] = append(edges, j)
	} else {
		graph.edges[i] = []int{j}
	}
}

func (graph *Graph[T]) AddNode(node T) {
	graph.nodes = append(graph.nodes, node)
}

/*
 * For lack of a better name, `Evaluate` processes the directed graph using a depth-first search.
 * Given a callback function, the function is called with the index of each graph's leaves. Then,
 * those nodes are pruned from the graph (although the graph is not modified) and the function is
 * called with the new leaves. This process is repeated until no leaves remain.
 *
 * Note that this function assumes that an edge from A to B indicates that A is a dependency of B.
 *
 * The function returns whether the graph is acyclic (i.e. whether every node was processed).
 */
func (graph *Graph[T]) Evaluate(evaluator func(i int)) bool {
	dependencyCount := make(map[int]int, len(graph.nodes))

	for _, dependents := range graph.edges {
		for _, dependent := range dependents {
			dependencyCount[dependent]++
		}
	}

	for i := range graph.nodes {
		if _, ok := dependencyCount[i]; !ok {
			dependencyCount[i] = 0
		}
	}

	stack := []int{}

	for i, n := range dependencyCount {
		if n == 0 {
			stack = append(stack, i)
		}
	}

	processed := 0

	for len(stack) > 0 {
		i := stack[len(stack)-1]

		stack = stack[:len(stack)-1]

		evaluator(i)

		for _, j := range graph.edges[i] {
			dependencyCount[j]--

			if dependencyCount[j] == 0 {
				stack = append(stack, j)
			}
		}

		processed++
	}

	return processed == len(graph.nodes)
}

func (graph *Graph[T]) GetNode(i int) T {
	return graph.nodes[i]
}

func (graph *Graph[T]) Length() int {
	return len(graph.nodes)
}

func NewGraph[T any]() *Graph[T] {
	return &Graph[T]{
		nodes: []T{},
		edges: map[int][]int{},
	}
}

func LinkedListToSlice[LinkedList any, Element any](
	linkedList *LinkedList,
	head func(*LinkedList) Element,
	tail func(*LinkedList) *LinkedList,
) ([]Element, *LinkedList) {
	result := []Element{}

	if linkedList == nil {
		return result, nil
	}

	current := linkedList

	for {
		result = append(result, head(current))
		next := tail(current)

		if next == nil {
			break
		}

		current = next
	}

	return result, current
}
