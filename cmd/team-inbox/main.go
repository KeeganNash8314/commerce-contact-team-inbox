package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/example/commerce-contact-inbox/internal/contact"
	"github.com/example/commerce-contact-inbox/internal/infrai"
)

type server struct {
	teamInbox string
	email     infrai.Client
}

func main() {
	teamInbox := os.Getenv("TEAM_INBOX")
	if teamInbox == "" || os.Getenv("INFRAI_API_KEY") == "" {
		log.Fatal("TEAM_INBOX and INFRAI_API_KEY are required")
	}
	s := server{teamInbox: teamInbox, email: infrai.Client{APIKey: os.Getenv("INFRAI_API_KEY")}}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /contact", s.handleContact)

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}
	log.Printf("team inbox service listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func (s server) handleContact(w http.ResponseWriter, r *http.Request) {
	var submission contact.Submission
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&submission); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid contact submission"})
		return
	}
	mail, err := contact.Route(s.teamInbox, submission)
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		return
	}

	result, err := s.email.Send(r.Context(), infrai.Email{To: mail.To, Subject: mail.Subject, Body: mail.Text}, "contact-"+submission.OrderID+"-"+submission.Kind)
	if err != nil {
		var apiErr *infrai.APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode >= 400 && apiErr.StatusCode < 500 {
			writeJSON(w, apiErr.StatusCode, map[string]string{"error": apiErr.Message})
			return
		}
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "email delivery failed"})
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"message_id": result.MessageID})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("encode response: %v", err)
	}
}
