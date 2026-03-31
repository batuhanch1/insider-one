package email_provider

type DeliverRequest struct {
	To      string `json:"to,omitempty"`
	From    string `json:"from,omitempty"`
	Subject string `json:"subject,omitempty"`
	Content string `json:"content,omitempty"`
}

func NewDeliverRequest(to, from, subject, content string) *DeliverRequest {
	return &DeliverRequest{
		To:      to,
		From:    from,
		Subject: subject,
		Content: content,
	}
}

type DeliverResponse struct {
	MessageID string `json:"message_id,omitempty"`
	Status    string `json:"status,omitempty"`
}

func (c *DeliverResponse) IsAccepted() bool {
	return c.Status == "accepted"
}
