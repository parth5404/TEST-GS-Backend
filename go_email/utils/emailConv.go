package utils

import (
	//"encoding/json"
	"bytes"
	"fmt"
	"html/template"
	"net/http"
)

func EmailConv(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/accountCreationTemplate.html")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Name string
	}{
		Name: "User", 
	}
	var emailBody string
	buf := &bytes.Buffer{}
	if err = tmpl.Execute(buf, data); err != nil {
		http.Error(w, "Template execution error", http.StatusInternalServerError)
		return
	}
	emailBody = buf.String()
	
	err = SendEmail("parthlahoti5404@gmail.com", "Sher", emailBody)
	if err != nil {
		http.Error(w, "Failed to send email", http.StatusInternalServerError)
		return
	}
	
	fmt.Fprintf(w, "Email sent successfully!")
	if err!= nil {
		http.Error(w, "Template execution error", http.StatusInternalServerError)
		return
	}
}