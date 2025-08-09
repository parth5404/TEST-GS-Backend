package main

import (
	"fmt"
	"net/http"
	"os"
	"log"
	"github.com/joho/godotenv"
	"github.com/parth5404/TEST-GS-Backend/go_email/utils"
)



func main() {
	if os.Getenv("ENVIRONMENT") == "development" {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}
	http.HandleFunc("/send-email", utils.EmailConv)

	port := os.Getenv("PORT")
	fmt.Println(port)
	fmt.Println("Server starting on port " + port + "...")
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
	//utils.SendEmail("parthlahoti5404@gmail.com", "Test Email", templates.AccountCreationTemplate("Parth Lahoti"))
}
