package poller

type DepthMessage struct {
	Stream string      `json:"stream"`
	Data   DepthUpdate `json:"data"`
}

// simplify response, remove unnecessary fields.
type DepthUpdate struct {
	Symbol string     `json:"s"`
	Bids   [][]string `json:"b"`
	Asks   [][]string `json:"a"`
}
