package http_client

// SMSRequest represents an SMS request payload
type SMSRequest struct {
	To      string `json:"to"`
	From    string `json:"from"`
	Message string `json:"message"`
}

// SMSResponse represents an SMS response
type SMSResponse struct {
	MessageID string `json:"message_id"`
	Status    string `json:"status"`
	Cost      string `json:"cost,omitempty"`
}
