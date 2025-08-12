package utils
import (
	"html/template"
	"os"
	"fmt"
	"bytes"
)

func GetTemplate(req EmailRequest) (string, error) {
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
	case "contactUsEmail":
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
	case "test":
		data = struct {
			Name string
		}{
			Name: req.FirstName + " " + req.LastName,
		}
	default:
		data = struct {
			Template string
		}{
			Template: req.Template,
		}
		fmt.Println(data)
	}
	var tmpl *template.Template
	var emailBody string
	if _, err := os.Stat("mail/templates/" + req.Template + ".html"); err == nil {
		tmpl, err = template.ParseFiles("mail/templates/" + req.Template + ".html")
		if err != nil {
			return "Template error", err
		}
		buf := &bytes.Buffer{}
		if err = tmpl.Execute(buf, data); err != nil {
			return "Template execution error", err
		}
		emailBody = buf.String()
	} else {
		emailBody = req.Template
	}
	return emailBody, nil
}