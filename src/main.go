package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"regexp"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

func handleSignalMessage(message *DataMessage, accountNumber string, sender string) {
	sendTypingIndicator("PUT", accountNumber, sender)
	responseText := getOpenWebUIResponse(message.Message, message.Attachments)
	sendTypingIndicator("DELETE", accountNumber, sender)
	sendSignalMessage(responseText, accountNumber, sender)
}

func main() {
	// Load our environment variables from '.env'
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file. Is it missing? Refer to the example .env file in the repository.")
		return
	}

	re := regexp.MustCompile(`^![a-z] *`)

	// Connect to Signal API WebSocket
	// TODO: refactor for more than one account.
	debug := os.Getenv("DEBUG")
	signalUrl := os.Getenv("SIGNAL_URL")
	signalNumber := os.Getenv("SIGNAL_NUMBER")
	defaultModel := os.Getenv("OPENWEBUI_MODEL_DEFAULT")
	// mise en place
	accountsMap := make(map[string]string)
	if _, err := os.Stat("accounts.json"); errors.Is(err, os.ErrNotExist) {
		log.Println("Accounts.json does not exist, creating and skipping read.")
		os.WriteFile("accounts.json", make([]byte, 0), 0660)
	} else {
		accountBytes, err := os.ReadFile("accounts.json")
		if err != nil {
			log.Println("Error opening file. ")
		}
		err = json.Unmarshal([]byte(accountBytes), &accountsMap)
	}

	modelsMap := make(map[string]string)
	if _, err := os.Stat("models.json"); errors.Is(err, os.ErrNotExist) {
		log.Println("Models.json does not exist, creating and skipping read.")
		os.WriteFile("models.json", make([]byte, 0), 0660)
	} else {
		modelBytes, err := os.ReadFile("models.json")
		if err != nil {
			log.Println("Error opening file. ")
		}
		err = json.Unmarshal([]byte(modelBytes), &modelsMap)
	}

	apiURL := url.URL{Scheme: "ws", Host: signalUrl, Path: "/v1/receive/" + signalNumber}
	conn, _, err := websocket.DefaultDialer.Dial(apiURL.String(), nil)
	if err != nil {
		log.Fatal("Failed to connect: ", err)
	}
	defer conn.Close()

	fmt.Println("Connected to Signal API. Waiting for messages...")

	// Read messages in a loop
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}

		var signalMessage SignalMessage
		if err := json.Unmarshal(message, &signalMessage); err != nil {
			log.Println("JSON parse error:", err)
			continue
		}

		// Extract and print just the message text
		if signalMessage.Envelope.DataMessage != nil {
			textMessage := signalMessage.Envelope.DataMessage.Message
			senderNumber := signalMessage.Envelope.SourceNumber
			if debug == "1" {
				fmt.Println("Text:", textMessage)
				fmt.Println("Message:", string(message))
			}

			match := re.FindString(textMessage)
			if debug == "1" {
				fmt.Println("Regex Result:", match)
			}

			if match == "" {
				if _, ok := accountsMap[signalMessage.Envelope.SourceNumber]; !ok {
					if debug == "1" {
						fmt.Println("New user, creating new chat.")
					}

					os.Setenv("OPENWEBUI_MODEL", defaultModel)
					newChatId := createNewChat(textMessage, signalNumber)
					accountsMap[senderNumber] = newChatId
					accountsJson, _ := json.Marshal(accountsMap)
					err = os.WriteFile("accounts.json", accountsJson, 0660)
					if err != nil {
						log.Println("Failed to create updated accounts.json. Check integrity of existing file.")
						log.Println("Then, check that this program has sufficient privileges to create files in the running directory.")
						pwd, _ := os.Getwd()
						log.Println("Current running directory: " + pwd)
						log.Println(err)
					}

					modelsMap[senderNumber] = defaultModel
					modelsJson, _ := json.Marshal(modelsMap)
					err = os.WriteFile("models.json", modelsJson, 0660)
					if err != nil {
						log.Println("Failed to update models.json. Check integrity of existing file.")
						log.Println("Then, check that this program has sufficient privileges to create files in the running directory.")
						pwd, _ := os.Getwd()
						log.Println("Current running directory: " + pwd)
						log.Println(err)
					}
				} else {
					os.Setenv("OPENWEBUI_CHAT_ID", accountsMap[senderNumber])
					modelBytes, err := os.ReadFile("models.json")
					if err != nil {
						log.Println("Error opening file. ")
					}
					err = json.Unmarshal([]byte(modelBytes), &modelsMap)
					currentModel := modelsMap[senderNumber]
					if currentModel == "" {
						currentModel = defaultModel
						modelsMap[senderNumber] = defaultModel
						modelsJson, _ := json.Marshal(modelsMap)
						err = os.WriteFile("models.json", modelsJson, 0660)
						if err != nil {
							log.Println("Failed to update models.json. Check integrity of existing file.")
							log.Println("Then, check that this program has sufficient privileges to create files in the running directory.")
							pwd, _ := os.Getwd()
							log.Println("Current running directory: " + pwd)
							log.Println(err)
						}
					}
					os.Setenv("OPENWEBUI_MODEL", modelsMap[senderNumber])
				}
				handleSignalMessage(signalMessage.Envelope.DataMessage, signalNumber, senderNumber)
			} else {
				sendTypingIndicator("PUT", signalNumber, senderNumber)
				responseText := parseCommand(textMessage, senderNumber)
				sendTypingIndicator("DELETE", signalNumber, senderNumber)
				sendSignalMessage(responseText, signalNumber, senderNumber)
			}
		}

		// Optionally write the entire received message to stdout for debugging
		if debug == "1" {
			fmt.Fprintln(os.Stdout, string(message))
			bufio.NewWriter(os.Stdout).Flush()
		}
	}
}

func parseCommand(textMessage, senderNumber string) string {
	commandVerb := textMessage[1]

	commandRegex := regexp.MustCompile(`\s(.*)`)
	command := commandRegex.FindString(textMessage)

	switch commandVerb {
	case 'm':
		return handleModelCommand(command, senderNumber)
	case 'w':
		return handleWebSearchCommand(command)
	default:
		return "Unknown command, nothing done."
	}

}
