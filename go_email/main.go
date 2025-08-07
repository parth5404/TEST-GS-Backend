package main

import (
	//"encoding/json"
	"fmt"
	//"html/template"
	"net/http"

	"github.com/parth5404/TEST-GS-Backend/go_email/utils"
	//"github.com/parth5404/TEST-GS-Backend/go_email/utils"
)



func main() {
	http.HandleFunc("/", utils.EmailConv)
	fmt.Println("Server starting on port 8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
	//utils.SendEmail("parthlahoti5404@gmail.com", "Test Email", templates.AccountCreationTemplate("Parth Lahoti"))
}
