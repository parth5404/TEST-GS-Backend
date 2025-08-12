package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"github.com/parth5404/TEST-GS-Backend/go_email/services"
)

type EmailRequest struct {
	FirstName string                 `json:"firstName"`
	LastName  string                 `json:"lastName"`
	Email     string                 `json:"email"`
	Subject   string                 `json:"subject"`
	Body      string                 `json:"body"`
	Template  string                 `json:"template"`
	ExtraData map[string]interface{} `json:"extraData"` // Use map if it can vary
}

type EmailData struct {
	Name      string
	ExtraData map[string]interface{} `json:"extraData"`
}

func EmailConv(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}
	var req EmailRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	fmt.Println(req)
	emailBody, err := GetTemplate(req)
	if err != nil {
		http.Error(w, "Failed to get template", http.StatusInternalServerError)
		return
	}
	err = services.SendEmail(req.Email, req.Subject, emailBody)
	if err != nil {
		http.Error(w, "Failed to send email", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Email sent successfully!")
}