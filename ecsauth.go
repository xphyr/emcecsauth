package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/go-resty/resty"
	"golang.org/x/crypto/ssh/terminal"
)

func main() {
	// Setup the basic command line arguments
	serverPtr := flag.String("server", "server.example.com", "ECS Cluster to Connect to")
	userNamePtr := flag.String("username", "user@example.com", "Username to authenticate as")
	verbosityPtr := flag.Bool("verbose", false, "Enable extra output for debugging.")
	listOnlyPtr := flag.Bool("listonly", false, "Only list current keys")
	expirationPtr := flag.Int("timeoutexpiration", 0, "expiration time in minutes (optional)")
	deactivatePtr := flag.Bool("deactivate", false, "deactivate all issued keys")
	flag.Parse()

	username, password := credentials(*userNamePtr)

	// GET request
	// Basic Auth for all request
	resty.SetBasicAuth(username, password)
	resty.RemoveProxy()

	if *serverPtr == "server.example.com" {
		fmt.Print("\nEnter Servername: ")
		reader := bufio.NewReader(os.Stdin)
		inputStr, _ := reader.ReadString('\n')
		inputStr = strings.TrimSpace(inputStr)
		flag.Set("server", inputStr)
	}

	reqBaseURL := "https://" + *serverPtr + ":4443"
	reqLoginURL := reqBaseURL + "/login"
	reqKeyURL := reqBaseURL + "/object/secret-keys"

	if *verbosityPtr == true {
		fmt.Printf("Username: %s\n", username)
		fmt.Println("Login URL: " + reqLoginURL)
	}

	resp, err := resty.R().Get(reqLoginURL)

	authToken := resp.Header()["X-Sds-Auth-Token"][0]

	if *verbosityPtr == true {
		// explore response object
		fmt.Printf("\nError: %v", err)
		fmt.Printf("\nResponse Status Code: %v", resp.StatusCode())
		fmt.Printf("\nResponse Status: %v", resp.Status())
		fmt.Printf("\nResponse Time: %v", resp.Time())
		fmt.Printf("\nResponse Recevied At: %v", resp.ReceivedAt())
		fmt.Println("\nRespone AuthToken: ", authToken)
		fmt.Printf("\nResponse Body: %v", resp.String()) // or resp.String() or string(resp.Body())
	}

	//lets try getting the current auth-Tokens
	resty.SetHeader("Accept", "application/json")
	resty.SetHeaders(map[string]string{
		"Content-Type":     "application/json",
		"X-SDS-AUTH-TOKEN": authToken,
	})

	resp, err = resty.R().Get(reqKeyURL)
	var respKey1 interface{}
	json.Unmarshal([]byte(resp.String()), &respKey1)

	if *verbosityPtr == true {
		// explore response object
		fmt.Printf("\nError: %v", err)
		fmt.Printf("\nResponse Status Code: %v", resp.StatusCode())
		fmt.Printf("\nResponse Status: %v", resp.Status())
		fmt.Printf("\nResponse Time: %v", resp.Time())
		fmt.Printf("\nResponse Recevied At: %v", resp.ReceivedAt())
		fmt.Println("\nRespone AuthToken: ", authToken)
		fmt.Printf("\nResponse Body: %v", resp.String()) // or resp.String() or string(resp.Body())
	}

	test := respKey1.(map[string]interface{})
	fmt.Println("\nHere are your current keys")
	fmt.Println("Secret Key 1: ", test["secret_key_1"])
	fmt.Println("Secret Key 1 Expiration: ", test["key_expiry_timestamp_1"])
	fmt.Println("Secret Key 2: ", test["secret_key_2"])
	fmt.Println("Secret Key 2 Expiration: ", test["key_expiry_timestamp_2"])

	if *listOnlyPtr == false {
		//generate request Body
		var reqBody string
		if *expirationPtr == 0 {
			reqBody = "{}"
		} else {
			reqBody = "{ \"existing_key_expiry_time_mins\": \"" + strconv.Itoa(*expirationPtr) + "\"}"
		}
		fmt.Println(reqBody)
		// generate a new keys
		resp, err = resty.R().
			SetBody(reqBody).
			Post(reqKeyURL)

		if *verbosityPtr == true {
			// explore response object
			fmt.Printf("\nError: %v", err)
			fmt.Printf("\nResponse Status Code: %v", resp.StatusCode())
			fmt.Printf("\nResponse Status: %v", resp.Status())
			fmt.Printf("\nResponse Time: %v", resp.Time())
			fmt.Printf("\nResponse Recevied At: %v", resp.ReceivedAt())
			fmt.Println("\nRespone AuthToken: ", authToken)
			fmt.Printf("\nResponse Body: %v", resp.String()) // or resp.String() or string(resp.Body())
		}
	}

	if *deactivatePtr == true {
		reqKeyDelURL := reqBaseURL + "/object/secret-keys/deactivate"
		reqBody := "{}"
		resp, err = resty.R().
			SetBody(reqBody).
			Post(reqKeyDelURL)

		if *verbosityPtr == true {
			// explore response object
			fmt.Printf("\nError: %v", err)
			fmt.Printf("\nResponse Status Code: %v", resp.StatusCode())
			fmt.Printf("\nResponse Status: %v", resp.Status())
			fmt.Printf("\nResponse Time: %v", resp.Time())
			fmt.Printf("\nResponse Recevied At: %v", resp.ReceivedAt())
			fmt.Println("\nRespone AuthToken: ", authToken)
			fmt.Printf("\nResponse Body: %v", resp.String()) // or resp.String() or string(resp.Body())
		}
	}
}

func credentials(username string) (string, string) {
	reader := bufio.NewReader(os.Stdin)
	password := ""

	if username == "user@example.com" {
		fmt.Print("Enter Username: ")
		username, _ = reader.ReadString('\n')
	}
	fmt.Print("Enter Password: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	if err == nil {
		password = string(bytePassword)
		fmt.Println("\n")
	} else {
		fmt.Println("Error capturing password")
	}

	return strings.TrimSpace(username), strings.TrimSpace(password)
}
