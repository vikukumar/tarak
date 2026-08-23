package mesh

// Destination represents a target service and version with weight.
type Destination struct {
	Service string            `json:"service"`
	Subset  string            `json:"subset"`
	Weight  int               `json:"weight"` // 1-100 percentage
	Tags    map[string]string `json:"tags,omitempty"`
}

// TrafficRoute defines intelligent routing, canary releases, and path matching.
type TrafficRoute struct {
	Name         string            `json:"name"`
	Namespace    string            `json:"namespace"`
	Host         string            `json:"host"`
	PathPrefix   string            `json:"pathPrefix,omitempty"`
	HeaderMatch  map[string]string `json:"headerMatch,omitempty"`
	Destinations []Destination     `json:"destinations"`
	TimeoutMs    int               `json:"timeoutMs,omitempty"`
	Retries      int               `json:"retries,omitempty"`
}
