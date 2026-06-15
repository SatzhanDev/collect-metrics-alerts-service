// Package audit реализует механизм аудита событий изменения метрик.
package audit

// Event описывает событие изменения метрик, фиксируемое системой аудита.
type Event struct {
	TS        int64    `json:"ts"`         // TS — Unix-время события.
	Metrics   []string `json:"metrics"`    // Metrics — имена изменённых метрик.
	IPAddress string   `json:"ip_address"` // IPAddress — IP-адрес клиента.
}
