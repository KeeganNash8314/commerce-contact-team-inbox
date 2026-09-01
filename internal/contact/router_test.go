package contact

import (
	"strings"
	"testing"
)

func TestRouteOrderContact(t *testing.T) {
	tests := []struct {
		name        string
		submission  Submission
		wantSubject string
		wantBody    string
	}{
		{"checkout", Submission{Kind: "checkout", OrderID: "ORD-1042", Customer: "Ada", Email: "ada@example.com", Message: "Card was declined"}, "Checkout question: ORD-1042", "Checkout needs review"},
		{"fulfillment", Submission{Kind: "fulfillment", OrderID: "ORD-1043", Customer: "Lin", Email: "lin@example.com", Message: "When will this ship?"}, "Fulfillment question: ORD-1043", "Fulfillment needs review"},
		{"receipt", Submission{Kind: "receipt", OrderID: "ORD-1044", Customer: "Sam", Email: "sam@example.com", Message: "Please resend it", AmountCents: 2599}, "Receipt question: ORD-1044", "Receipt amount: $25.99"},
		{"order update", Submission{Kind: "order_update", OrderID: "ORD-1045", Customer: "Mina", Email: "mina@example.com", Message: "Change the address"}, "Order update: ORD-1045", "Customer requested an order update"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mail, err := Route("orders@example.com", tt.submission)
			if err != nil {
				t.Fatalf("Route() error = %v", err)
			}
			if mail.To != "orders@example.com" || mail.Subject != tt.wantSubject || !strings.Contains(mail.Text, tt.wantBody) {
				t.Fatalf("Route() = %#v", mail)
			}
		})
	}
}

func TestRouteRejectsUnknownKind(t *testing.T) {
	_, err := Route("orders@example.com", Submission{Kind: "return", OrderID: "ORD-9", Customer: "Jo", Email: "jo@example.com"})
	if err == nil {
		t.Fatal("Route() accepted an unknown kind")
	}
}
