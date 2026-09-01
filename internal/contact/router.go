package contact

import (
	"fmt"
	"strings"
)

type Submission struct {
	Kind        string `json:"kind"`
	OrderID     string `json:"order_id"`
	Customer    string `json:"customer"`
	Email       string `json:"email"`
	Message     string `json:"message"`
	AmountCents int    `json:"amount_cents,omitempty"`
}

type Mail struct {
	To      string
	Subject string
	Text    string
}

func Route(teamInbox string, s Submission) (Mail, error) {
	if strings.TrimSpace(teamInbox) == "" {
		return Mail{}, fmt.Errorf("team inbox is required")
	}
	if strings.TrimSpace(s.OrderID) == "" || strings.TrimSpace(s.Customer) == "" || strings.TrimSpace(s.Email) == "" {
		return Mail{}, fmt.Errorf("order_id, customer, and email are required")
	}

	var subject, detail string
	switch s.Kind {
	case "checkout":
		subject = "Checkout question"
		detail = "Checkout needs review"
	case "fulfillment":
		subject = "Fulfillment question"
		detail = "Fulfillment needs review"
	case "receipt":
		subject = "Receipt question"
		detail = fmt.Sprintf("Receipt amount: $%.2f", float64(s.AmountCents)/100)
	case "order_update":
		subject = "Order update"
		detail = "Customer requested an order update"
	default:
		return Mail{}, fmt.Errorf("kind must be checkout, fulfillment, receipt, or order_update")
	}

	return Mail{
		To:      teamInbox,
		Subject: fmt.Sprintf("%s: %s", subject, s.OrderID),
		Text: fmt.Sprintf("%s\nOrder: %s\nCustomer: %s <%s>\nMessage: %s",
			detail, s.OrderID, s.Customer, s.Email, s.Message),
	}, nil
}
