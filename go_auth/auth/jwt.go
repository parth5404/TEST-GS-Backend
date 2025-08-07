package auth

import (
	"fmt"
	"os"
)

// secret is declared at package level since it will likely be used by multiple functions
var secret = os.Getenv("JWT_SECRET")

func Jwt() {
	fmt.Println(secret)
}
