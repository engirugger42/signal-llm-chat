package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type ModelsResponse struct {
	Models []Model `json:"models"`
}

type Model struct {
	Name           string       `json:"name"`
	Model          string       `json:"model"`
	ModifiedAt     string       `json:"modified_at"`
	Size           int64        `json:"size"`
	Digest         string       `json:"digest"`
	Details        ModelDetails `json:"details"`
	ConnectionType string       `json:"connection_type"`
	URLs           []int        `json:"urls"`
}

type ModelDetails struct {
	ParentModel       string    `json:"parent_model"`
	Format            string    `json:"format"`
	Family            string    `json:"family"`
	Families          *[]string `json:"families"`
	ParameterSize     string    `json:"parameter_size"`
	QuantizationLevel string    `json:"quantization_level"`
}

func sendOllamaCommand(verb, command string, payload []byte) []byte {
	apikey := os.Getenv("OPENWEBUI_API_KEY")
	url := os.Getenv("OPENWEBUI_URL")

	req, err := http.NewRequest(
		verb,
		"http://"+url+"/ollama/api/"+command,
		bytes.NewBuffer(payload),
	)

	if err != nil {
		log.Println("Failed to send Ollama request:", err)
	}

	req.Header.Set("Authorization", "Bearer "+apikey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
	}
	defer resp.Body.Close()

	fmt.Println("Response status:", resp.Status)
	body, err := io.ReadAll(resp.Body)

	return body
}

func handleModelChangeCommand(model, senderNumber string) string {
	modelsMap := make(map[string]string)
	modelBytes, err := os.ReadFile("models.json")
	if err != nil {
		log.Println("Error opening file. ")
	}
	err = json.Unmarshal([]byte(modelBytes), &modelsMap)
	modelsMap[senderNumber] = model
	modelsJson, _ := json.Marshal(modelsMap)
	err = os.WriteFile("models.json", modelsJson, 0660)
	if err != nil {
		log.Println("Failed to update models.json. Check integrity of existing file.")
		log.Println("Then, check that this program has sufficient privileges to create files in the running directory.")
		pwd, _ := os.Getwd()
		log.Println("Current running directory: " + pwd)
		log.Println(err)
	}

	err = os.Setenv("OPENWEBUI_MODEL", model)

	if err != nil {
		fmt.Println("Error setting model variable: ", err)
		return "Failed to set model, check server logs for details."
	}

	return "Model set to " + model
}

func handleModelListCommand(command string) string {
	body := sendOllamaCommand("GET", command, nil)
	var response ModelsResponse
	err := json.Unmarshal(body, &response)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
	}

	var modelList []string
	for _, model := range response.Models {
		modelList = append(modelList, model.Name)
	}

	modelListString := strings.Join(modelList, "\n")

	return modelListString
}

func handleWebSearchCommand(command string) string {
	commandElements := strings.Fields(command)

	if len(commandElements) == 0 {
		return toggleWebSearch()
	}

	switch commandElements[0] {
	case "true":
		fallthrough
	case "1":
		fallthrough
	case "on":
		os.Setenv("OPENWEBUI_WEB_SEARCH", "1")
		return "Web search enabled."
	case "false":
		fallthrough
	case "0":
		fallthrough
	case "off":
		os.Setenv("OPENWEBUI_WEB_SEARCH", "1")
		return "Web search enabled."
	default:
		return toggleWebSearch()
	}
}

func toggleWebSearch() string {
	currentState := os.Getenv("OPENWEBUI_WEB_SEARCH")
	if os.Getenv("DEBUG") == "1" {
		fmt.Println("Current value of web search: ", currentState)
	}
	if currentState == "1" {
		os.Setenv("OPENWEBUI_WEB_SEARCH", "0")
		return "Web search disabled."
	} else {
		os.Setenv("OPENWEBUI_WEB_SEARCH", "1")
		return "Web search enabled."
	}
}

func handleModelCommand(command, senderNumber string) string {
	commandElements := strings.Fields(command)

	if len(commandElements) == 0 {
		modelsMap := make(map[string]string)
		modelBytes, err := os.ReadFile("models.json")
		if err != nil {
			log.Println("Error opening file. ")
		}
		err = json.Unmarshal([]byte(modelBytes), &modelsMap)

		return "Your current model is " + modelsMap[senderNumber]
	}

	switch commandElements[0] {
	case "list":
		return handleModelListCommand("tags")
	case "load":
		return handleModelChangeCommand(commandElements[1], senderNumber)
	default:
		return command + " not implemented at this time."
	}
}
