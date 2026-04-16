package src

type SearchOutput struct {
	Matches      []*Node
	TraversalLog []*Node
	Count int
	TraversalTime int64
}
