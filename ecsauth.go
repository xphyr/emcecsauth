package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/go-resty/resty"
	"golang.org/x/crypto/ssh/terminal"
)

func main() {
	// Setup the basic command line arguments
	serverPtr := flag.String("server", "server.example.com", "ECS Cluster to Connect to")
	userNamePtr := flag.String("username", "user@example.com", "Username to authenticate as")
	flag.Parse()

	username, password := credentials(*userNamePtr)
	fmt.Printf("Username: %s\n", username)

	// GET request
	// Basic Auth for all request
	resty.SetBasicAuth(username, password)
	resty.RemoveProxy()

	reqBaseURL := "https://" + *serverPtr + ":4443"
	reqLoginURL := reqBaseURL + "/login"
	reqKeyURL := reqBaseURL + "/object/secret-keys"
	fmt.Println("Login URL: " + reqLoginURL)
	resp, err := resty.R().Get(reqLoginURL)

	authToken := resp.Header()["X-Sds-Auth-Token"][0]

	// explore response object
	// fmt.Printf("\nError: %v", err)
	// fmt.Printf("\nResponse Status Code: %v", resp.StatusCode())
	// fmt.Printf("\nResponse Status: %v", resp.Status())
	// fmt.Printf("\nResponse Time: %v", resp.Time())
	// fmt.Printf("\nResponse Recevied At: %v", resp.ReceivedAt())
	fmt.Println("\nRespone AuthToken: ", authToken)
	// fmt.Printf("\nResponse Body: %v", resp.String()) // or resp.String() or string(resp.Body())

	//lets try getting the current auth-Tokens
	resty.SetHeader("Accept", "application/json")
	resty.SetHeaders(map[string]string{
		"Content-Type":     "application/json",
		"X-SDS-AUTH-TOKEN": authToken,
	})

	resp, err = resty.R().Get(reqKeyURL)
	// explore response object
	fmt.Printf("\nError: %v", err)
	fmt.Printf("\nResponse Status Code: %v", resp.StatusCode())
	fmt.Printf("\nResponse Status: %v", resp.Status())
	fmt.Printf("\nResponse Time: %v", resp.Time())
	fmt.Printf("\nResponse Recevied At: %v", resp.ReceivedAt())
	fmt.Printf("\nResponse Body: %v", resp.String()) // or resp.String() or string(resp.Body())

}

func credentials(username string) (string, string) {
	reader := bufio.NewReader(os.Stdin)
	password := ""

	if username == "" {
		fmt.Print("Enter Username: ")
		username, _ = reader.ReadString('\n')
	}
	fmt.Print("Enter Password: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	if err == nil {
		password = string(bytePassword)
	} else {
		fmt.Println("Error capturing password")
	}

	return strings.TrimSpace(username), strings.TrimSpace(password)
}
