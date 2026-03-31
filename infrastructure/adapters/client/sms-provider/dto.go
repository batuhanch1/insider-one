package sms_provider

type DeliverRequest struct {
	PhoneNumber string `json:"phone_number,omitempty"`
	Sender      string `json:"sender,omitempty"`
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
	Status    string `json:"status"`
	MessageID string `json:"message_id"`
}

func (c *DeliverResponse) IsAccepted() bool {
	return c.Status == "accepted"
}
