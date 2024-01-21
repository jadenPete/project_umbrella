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

/*
 * `ConsolidatedGraph` is like `DirectedGraph`, but it's built to support an operation I call
 * consolidation (hence the name). Recall that `DirectedGraph` is mainly used to store block graphs.
 * When evaluating block graphs, we sometimes encounter dependency cycles between functions.
 * Consider the following code.
 * ```
 * fn foo():
 *     bar()
 *
 * fn bar():
 *     foo()
 * ```
 *
 * Normally, it'd be impossible to evaluate this block graph because `foo` depends on `bar` and
 * vice versa. To resolve this conundrum, we "consolidate" the nodes (functions) that comprise this
 * dependency cycle into a single node, transforming the graph from something like this:
 * ```
 * {
 *     "foo": ["bar"],
 *     "bar": ["foo"]
 * }
 * ````
 *
 * into something like this:
 * ```
 * {
 *     ("foo", "bar"): [("foo", "bar")]
 * }
 * ```
 *
 * Then, because self-references are permitted (for the same reason recursive functions work), this
 * block graph can be evaluated. This consolidation operation can be performed using the
 * `ConsolidatedGraph#Consolidate` method.
 */
type ConsolidatedGraph[T any] struct {
	nodes                   []T
	consolidatedNodeIndices []int
	dependencies            *DirectedGraph[*ConsolidatedGraphNode[T]]
	reverseDependencies     *DirectedGraph[*ConsolidatedGraphNode[T]]
}

func (graph *ConsolidatedGraph[T]) AddEdge(i int, j int) {
	iConsolidated := graph.consolidatedNodeIndices[i]
	jConsolidated := graph.consolidatedNodeIndices[j]

	graph.dependencies.AddEdge(iConsolidated, jConsolidated)
	graph.reverseDependencies.AddEdge(jConsolidated, iConsolidated)
}

func (graph *ConsolidatedGraph[T]) AddNode(node T) int {
	graph.nodes = append(graph.nodes, node)

	nodeIndex := len(graph.nodes) - 1
	consolidatedNode := &ConsolidatedGraphNode[T]{
		Consolidated:   nil,
		Unconsolidated: nodeIndex,
		IsConsolidated: false,
	}

	consolidatedNodeIndex := graph.dependencies.AddNode(consolidatedNode)

	graph.reverseDependencies.AddNode(consolidatedNode)

	graph.consolidatedNodeIndices = append(graph.consolidatedNodeIndices, consolidatedNodeIndex)

	return nodeIndex
}

/*
 * Transform the graph such that cycles are repeatedly "consolidated", which entails combining
 * cycles' nodes into larger nodes and updating the graph's dependencies and reverse dependencies
 * accordingly.
 *
 * Because the cycle's nodes' dependencies/reverse dependencies will now become the
 * consolidated node's dependencies/reverse dependencies, the consolidated node will depend on
 * itself, hence why cycles of length one are excluded when searching for cycles
 * (see the `ConsolidatedGraph#nextCycle` method).
 *
 * NOTE: As with `DirectedGraph`, an edge from node A to node B is interpreted as B depending on A.
 */
func (graph *ConsolidatedGraph[T]) Consolidate() {
	for {
		cycle, ok := graph.nextCycle()

		if !ok {
			break
		}

		consolidatedCycle := &ConsolidatedGraphNode[T]{
			Consolidated:   []int{},
			Unconsolidated: 0,
			IsConsolidated: true,
		}

		for _, i := range cycle {
			consolidatedCycle.Consolidated =
				append(consolidatedCycle.Consolidated, graph.dependencies.GetNode(i).Nodes()...)
		}

		consolidatedCycleIndex := graph.dependencies.AddNode(consolidatedCycle)

		graph.reverseDependencies.AddNode(consolidatedCycle)

		for _, i := range cycle {
			for j := range graph.dependencies.GetNode(i).Nodes() {
				graph.consolidatedNodeIndices[j] = consolidatedCycleIndex
			}

			/*
			 * Step 1: Make the current cycle node's reverse dependencies the consolidated nodes'.
			 *
			 * Note that this step is performed before Step 4 because if the current cycle node is
			 * self-referring, Step 2 will replace its dependency with itself with one to the
			 * consolidated node; if Step 4 attempts to replace the corresponding reverse dependency
			 * before it's been added here, it'll fail.
			 */
			for _, j := range graph.reverseDependencies.GetEdgesFrom(i) {
				graph.reverseDependencies.AddEdge(consolidatedCycleIndex, j)
			}

			/*
			 * Step 2: Make the current cycle node's dependents the consolidated nodes'
			 */
			for _, j := range graph.reverseDependencies.GetEdgesFrom(i) {
				graph.dependencies.AddEdge(j, consolidatedCycleIndex)
				graph.dependencies.RemoveEdge(j, i)
			}

			/*
			 * Step 3: Make the current cycle node's dependencies the consolidated nodes'.
			 *
			 * Note that this step is performed after Step 2 because if the current cycle node is
			 * self-referring, that dependency needs to be replaced with one to the
			 * consolidated node first so the consolidated node doesn't depend on the
			 * current cycle node.
			 */
			for _, j := range graph.dependencies.GetEdgesFrom(i) {
				graph.dependencies.AddEdge(consolidatedCycleIndex, j)
			}

			/*
			 * Step 4: Make the current cycle node's dependencies
			 * (as represented in `reverse_dependencies`) the consolidated nodes'
			 */
			for _, j := range graph.dependencies.GetEdgesFrom(i) {
				graph.reverseDependencies.AddEdge(j, consolidatedCycleIndex)
				graph.reverseDependencies.RemoveEdge(j, i)
			}

			graph.dependencies.RemoveNode(i)
			graph.reverseDependencies.RemoveNode(i)
		}
	}
}

// Like `DirectedGraph#Evaluate`, but yields consolidated nodes in addition to individual nodes.
func (graph *ConsolidatedGraph[T]) Evaluate(
	evaluator func(consolidatedNode *ConsolidatedGraphNode[T]),
) bool {
	return graph.dependencies.Evaluate(
		func(i int) {
			evaluator(graph.dependencies.GetNode(i))
		},
	)
}

func (graph *ConsolidatedGraph[T]) GetNode(i int) T {
	return graph.nodes[i]
}

func (graph *ConsolidatedGraph[T]) Length() int {
	return len(graph.nodes)
}

/*
 * Return the next cycle in the graph, excluding self-referential cycles
 * (those involving one node depending on itself).
 */
func (graph *ConsolidatedGraph[T]) nextCycle() ([]int, bool) {
	currentPath := []int{}
	currentPathIndices := map[int]int{}
	seen := map[int]interface{}{}

	type WithConsolidatedNode func(int, WithConsolidatedNode) ([]int, bool)

	withConsolidatedNode := func(i int, self WithConsolidatedNode) ([]int, bool) {
		if _, ok := seen[i]; ok {
			return nil, false
		}

		if j, ok := currentPathIndices[i]; ok {
			if j == len(currentPath)-1 {
				return nil, false
			}

			return currentPath[j:], true
		}

		currentPath = append(currentPath, i)
		currentPathIndices[i] = len(currentPath) - 1

		defer func() {
			currentPath = currentPath[:len(currentPath)-1]

			delete(currentPathIndices, i)

			seen[i] = nil
		}()

		for _, dependency := range graph.reverseDependencies.GetEdgesFrom(i) {
			if result, ok := self(dependency, self); ok {
				return result, true
			}
		}

		return nil, false
	}

	for _, i := range graph.dependencies.Nodes() {
		if result, ok := withConsolidatedNode(i, withConsolidatedNode); ok {
			return result, true
		}
	}

	return nil, false
}

func (graph *ConsolidatedGraph[T]) Nodes() []T {
	return graph.nodes
}

func (graph *ConsolidatedGraph[T]) RemoveConsolidatedSelfReferences(
	shouldRemove func(*ConsolidatedGraphNode[T]) bool,
) {
	for _, i := range graph.dependencies.Nodes() {
		if graph.dependencies.HasEdge(i, i) && shouldRemove(graph.dependencies.GetNode(i)) {
			graph.dependencies.RemoveEdge(i, i)
			graph.reverseDependencies.RemoveEdge(i, i)
		}
	}
}

func NewConsolidatedGraph[T any]() *ConsolidatedGraph[T] {
	return &ConsolidatedGraph[T]{
		nodes:                   []T{},
		consolidatedNodeIndices: []int{},
		dependencies:            NewDirectedGraph[*ConsolidatedGraphNode[T]](),
		reverseDependencies:     NewDirectedGraph[*ConsolidatedGraphNode[T]](),
	}
}

type ConsolidatedGraphNode[T any] struct {
	Consolidated   []int
	Unconsolidated int
	IsConsolidated bool
}

func (node *ConsolidatedGraphNode[T]) Nodes() []int {
	if node.IsConsolidated {
		return node.Consolidated
	}

	return []int{node.Unconsolidated}
}

type DirectedGraph[T any] struct {
	nodes    map[int]T
	nextNode int
	edges    map[int]map[int]interface{}
}

func (graph *DirectedGraph[T]) AddEdge(i int, j int) {
	if edges, ok := graph.edges[i]; ok {
		edges[j] = nil
	} else {
		graph.edges[i] = map[int]interface{}{
			j: nil,
		}
	}
}

func (graph *DirectedGraph[T]) AddNode(node T) int {
	graph.nodes[graph.nextNode] = node
	graph.nextNode++

	return graph.nextNode - 1
}

/*
 * For lack of a better name, `Evaluate` processes the directed graph using a depth-first search.
 * Given a callback function, the function is called with the index of each graph's leaves. Then,
 * those nodes are pruned from the graph (although the graph is not modified) and the function is
 * called with the new leaves. This process is repeated until no leaves remain.
 *
 * NOTE: An edge from node A to node B is interpreted as A depending on B.
 *
 * The function returns whether the graph is acyclic (i.e. whether every node was processed).
 */
func (graph *DirectedGraph[T]) Evaluate(evaluator func(i int)) bool {
	dependencyCount := make(map[int]int, len(graph.nodes))

	for _, dependents := range graph.edges {
		for dependent := range dependents {
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

		for j := range graph.edges[i] {
			dependencyCount[j]--

			if dependencyCount[j] == 0 {
				stack = append(stack, j)
			}
		}

		processed++
	}

	return processed == len(graph.nodes)
}

func (graph *DirectedGraph[T]) GetEdgesFrom(i int) []int {
	if edges, ok := graph.edges[i]; ok {
		result := []int{}

		for j := range edges {
			result = append(result, j)
		}

		return result
	}

	return []int{}
}

func (graph *DirectedGraph[T]) GetNode(i int) T {
	return graph.nodes[i]
}

func (graph *DirectedGraph[T]) HasEdge(i int, j int) bool {
	_, ok := graph.edges[i][j]

	return ok
}

func (graph *DirectedGraph[T]) Length() int {
	return len(graph.nodes)
}

func (graph *DirectedGraph[T]) Nodes() []int {
	result := []int{}

	for i := range graph.nodes {
		result = append(result, i)
	}

	return result
}

func (graph *DirectedGraph[T]) RemoveEdge(i int, j int) {
	delete(graph.edges[i], j)

	if len(graph.edges[i]) == 0 {
		delete(graph.edges, i)
	}
}

func (graph *DirectedGraph[T]) RemoveNode(i int) {
	delete(graph.nodes, i)
	delete(graph.edges, i)
}

func NewDirectedGraph[T any]() *DirectedGraph[T] {
	return &DirectedGraph[T]{
		nodes:    map[int]T{},
		nextNode: 0,
		edges:    map[int]map[int]interface{}{},
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
