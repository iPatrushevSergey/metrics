package audit

//go:generate easyjson -all $GOFILE

// Event is a single audit event for observers.
type Event struct {
	TS        int64    `json:"ts"`
	Metrics   []string `json:"metrics"`
	IPAddress string   `json:"ip_address"`
}
