package push_provider

type DeliverRequest struct {
	Sender      string `json:"sender,omitempty"`
	PhoneNumber string `json:"phone_number,omitempty"`
	Content     string `json:"content,omitempty"`
}

func NewDeliverRequest(sender, phoneNumber, content string) *DeliverRequest {
	return &DeliverRequest{
		Sender:      sender,
		PhoneNumber: phoneNumber,
		Content:     content,
	}
}

type DeliverResponse struct {
	MessageID string `json:"message_id"`
	Status    string `json:"status"`
}

func (c *DeliverResponse) IsAccepted() bool {
	return c.Status == "accepted"
}
