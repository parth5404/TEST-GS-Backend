package utils

import (
	//"encoding/json"
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
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
	extraJSON, err := json.Marshal(req.ExtraData)
	if err != nil {
		http.Error(w, "Failed to marshal extraData", http.StatusInternalServerError)
		return
	}
	extraJSONString := string(extraJSON)
	fmt.Println(extraJSONString)

	var data interface{}
	switch req.Template {
	case "accountCreationTemplate":
		data = struct {
			Name string
		}{
			Name: req.FirstName + " " + req.LastName,
		}
		fmt.Println(data)
	case "paymentSuccessEmailTemplate":
		data = struct {
			Name      string
			Amount    string `json:"amount"`
			OrderId   string `json:"orderId"`
			PaymentId string `json:"paymentId"`
		}{
			Name:      req.FirstName + " " + req.LastName,
			Amount:    fmt.Sprintf("%v", req.ExtraData["amount"]),
			OrderId:   fmt.Sprintf("%v", req.ExtraData["orderId"]),
			PaymentId: fmt.Sprintf("%v", req.ExtraData["paymentId"]),
		}
		fmt.Println(data)
	case "courseEnrollmentEmailTemplate":
		data = struct {
			CourseName string `json:"courseName"`
			Name       string `json:"name"`
		}{
			CourseName: fmt.Sprintf("%v", req.ExtraData["courseName"]),
			Name:       req.FirstName + " " + req.LastName,
		}
		fmt.Println(data)
	case "emailOtpTemplate":
		var otpStr string
		switch otp := req.ExtraData["otp"].(type) {
		case string:
			otpStr = otp
		case float64:
			otpStr = fmt.Sprintf("%.0f", otp)
		case int:
			otpStr = fmt.Sprintf("%d", otp)
		default:
			otpStr = fmt.Sprintf("%v", otp)
		}
		data = struct {
			OTP string `json:"otp"`
		}{
			OTP: otpStr,
		}
		fmt.Println(data)
	case "passwordUpdateTemplate":
		data = struct {
			Email string `json:"email"`
			Name  string `json:"name"`
		}{
			Email: req.Email,
			Name:  req.FirstName + " " + req.LastName,
		}
		fmt.Println(data)
	case "contactFormRes":
		data = struct {
			Email       string `json:"email"`
			FirstName   string `json:"firstName"`
			LastName    string `json:"lastName"`
			Message     string `json:"message"`
			PhoneNo     string `json:"phoneNo"`
			CountryCode string `json:"countryCode"`
		}{
			Email:       req.Email,
			FirstName:   req.FirstName,
			LastName:    req.LastName,
			Message:     fmt.Sprintf("%v", req.ExtraData["message"]),
			PhoneNo:     fmt.Sprintf("%v", req.ExtraData["phoneNo"]),
			CountryCode: fmt.Sprintf("%v", req.ExtraData["countryCode"]),
		}
		fmt.Println(data)
	default:
		data = struct {
			Name string
		}{
			Name: req.FirstName + " " + req.LastName,
		}
		fmt.Println(data)
	}
	var tmpl *template.Template
	var emailBody string
	if req.Body == "" {
		tmpl, err = template.ParseFiles("mail/templates/" + req.Template + ".html")
		if err != nil {
			http.Error(w, "Template error", http.StatusInternalServerError)
			return
		}
		buf := &bytes.Buffer{}
		if err = tmpl.Execute(buf, data); err != nil {
			http.Error(w, "Template execution error", http.StatusInternalServerError)
			return
		}
		emailBody = buf.String()
	} else {
		emailBody = req.Body
	}

	err = SendEmail(req.Email, req.Subject, emailBody)
	if err != nil {
		http.Error(w, "Failed to send email", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Email sent successfully!")
}
