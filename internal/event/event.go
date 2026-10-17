package event

type Item struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type OrderPlacedPayload struct {
	OrderID    string `json:"order_id"`
	CustomerID string `json:"customer_id"`
	Items      []Item `json:"items"`
}

type DomainEvent struct {
	EventID   string             `json:"event_id"`
	EventType string             `json:"event_type"`
	Timestamp string             `json:"timestamp"`
	Version   string             `json:"version"`
	Producer  string             `json:"producer"`
	Data      OrderPlacedPayload `json:"data"`
}
