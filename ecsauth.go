package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"syscall"

	"io/ioutil"

	"github.com/go-resty/resty"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	Trace   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

func InitLog(
	traceHandle io.Writer,
	infoHandle io.Writer,
	warningHandle io.Writer,
	errorHandle io.Writer) {

	Trace = log.New(traceHandle,
		"TRACE: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Info = log.New(infoHandle,
		"INFO: ", log.LstdFlags)

	Warning = log.New(warningHandle,
		"WARNING: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Error = log.New(errorHandle,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)
}

func main() {
	// Setup the basic command line arguments
	serverPtr := flag.String("server", "server.example.com", "ECS Cluster to Connect to")
	userNamePtr := flag.String("username", "user@example.com", "Username to authenticate as")
	verbosityPtr := flag.Bool("verbose", false, "Enable extra output for debugging.")
	// listOnlyPtr := flag.Bool("listonly", false, "Only list current keys")
	// expirationPtr := flag.Int("timeoutexpiration", 0, "expiration time in minutes (optional)")
	deactivatePtr := flag.Bool("deactivate", false, "deactivate all issued keys")
	flag.Parse()

	if *verbosityPtr == true {
		InitLog(os.Stdout, os.Stdout, os.Stdout, os.Stderr)
	} else {
		InitLog(ioutil.Discard, ioutil.Discard, os.Stdout, os.Stderr)
	}

	username, password := credentials(*userNamePtr)

	if *serverPtr == "server.example.com" {
		fmt.Print("Enter Servername: ")
		reader := bufio.NewReader(os.Stdin)
		inputStr, _ := reader.ReadString('\n')
		inputStr = strings.TrimSpace(inputStr)
		flag.Set("server", inputStr)
	}

	reqBaseURL := "https://" + *serverPtr + ":4443"

	authToken := serverLogin(username, password, reqBaseURL)

	//lets try getting the current auth-Tokens
	getS3AuthTokens(authToken, reqBaseURL)

	if *deactivatePtr == true {
		deleteS3Tokens(authToken, reqBaseURL)
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
		fmt.Print("\n")
	} else {
		Error.Fatalf("Error capturing password.")
	}

	return strings.TrimSpace(username), strings.TrimSpace(password)
}

func serverLogin(username string, password string, baseURL string) string {
	// GET request
	// Basic Auth for all request
	resty.SetBasicAuth(username, password)
	resty.RemoveProxy()

	reqLoginURL := baseURL + "/login"

	Info.Println("Username: " + username)
	Info.Println("Login URL: " + reqLoginURL)

	resp, err := resty.R().Get(reqLoginURL)
	if err != nil {
		Error.Fatalf("\n - Error connecting to ECS: %s", err)
	}

	authToken := resp.Header()["X-Sds-Auth-Token"][0]

	if authToken == "" {
		Error.Fatalln("ECS did not return an authToken.")
	}

	Info.Println("\nResponse Status Code: " + string(resp.StatusCode()))
	Info.Println("\nResponse Status: " + resp.Status())
	Info.Println("\nRespone AuthToken: " + authToken)
	Info.Println("\nResponse Body: " + resp.String()) // or resp.String() or string(resp.Body())

	return authToken
}

func deleteS3Tokens(authToken string, baseURL string) {
	reqKeyDelURL := baseURL + "/object/secret-keys/deactivate"
	reqBody := "{}"
	resp, _ := resty.R().
		SetBody(reqBody).
		Post(reqKeyDelURL)

	Info.Println("\nResponse Status Code: " + string(resp.StatusCode()))
	Info.Println("\nResponse Status: " + resp.Status())
	Info.Println("\nRespone AuthToken: " + authToken)
	Info.Println("\nResponse Body: " + resp.String()) // or resp.String() or string(resp.Body())

}

func getS3AuthTokens(authToken string, baseURL string) {
	resty.SetHeader("Accept", "application/json")
	resty.SetHeaders(map[string]string{
		"Content-Type":     "application/json",
		"X-SDS-AUTH-TOKEN": authToken,
	})
	reqKeyURL := baseURL + "/object/secret-keys"

	resp, _ := resty.R().Get(reqKeyURL)

	Info.Println("\nResponse Status Code: " + string(resp.StatusCode()))
	Info.Println("\nResponse Status: " + resp.Status())
	Info.Println("\nRespone AuthToken: " + authToken)
	Info.Println("\nResponse Body: " + resp.String()) // or resp.String() or string(resp.Body())

	var respKey1 interface{}
	json.Unmarshal([]byte(resp.String()), &respKey1)

	test := respKey1.(map[string]interface{})
	fmt.Println("\nHere are your current keys")
	fmt.Println("Secret Key 1: ", test["secret_key_1"])
	fmt.Println("Secret Key 1 Expiration: ", test["key_expiry_timestamp_1"])
	fmt.Println("Secret Key 2: ", test["secret_key_2"])
	fmt.Println("Secret Key 2 Expiration: ", test["key_expiry_timestamp_2"])

}
