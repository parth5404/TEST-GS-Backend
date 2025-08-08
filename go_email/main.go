package main

import (
	//"encoding/json"
	"fmt"
	//"html/template"
	"net/http"
	"os"
	"github.com/parth5404/TEST-GS-Backend/go_email/utils"
	//"github.com/parth5404/TEST-GS-Backend/go_email/utils"
)



func main() {
	http.HandleFunc("/send-email", utils.EmailConv)

	port := os.Getenv("PORT")
	fmt.Println("Server starting on port " + port + "...")
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
	//utils.SendEmail("parthlahoti5404@gmail.com", "Test Email", templates.AccountCreationTemplate("Parth Lahoti"))
}
