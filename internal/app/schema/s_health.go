package schema

type HealthLive struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

type HealthStatus struct {
	Status string `json:"status"`
}

type HealthReady struct {
	Status  string                  `json:"status"`
	Info    map[string]HealthStatus `json:"info"`
	Error   map[string]HealthStatus `json:"error"`
	Details map[string]HealthStatus `json:"details"`
}
