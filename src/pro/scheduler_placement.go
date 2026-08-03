package pro

// PlacementStrategy is the interface implemented by deployment placement
// strategies in Cosmos Pro. The free build ships no implementations; the
// scheduler entry points return early before any strategy is consulted.
type PlacementStrategy interface {
	Choose(in PlacementInput) []string
}

type PlacementInput struct {
	Deployment    Deployment
	EligibleNodes []string
	CurrentlyOn   []string
	NodeMetrics   map[string]NodeMetrics
}

type NodeMetrics struct {
	CPUPercent   float64
	RAMPercent   float64
	MonitoringOn bool
}
