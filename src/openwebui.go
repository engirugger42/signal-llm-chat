package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

type Data struct {
	Status string `json:"status"`
}

type OpenWebUIFile struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type OpenWebUIMessage struct {
	ID        string   `json:"id"`
	Role      string   `json:"role"`
	Content   string   `json:"content"`
	Timestamp int64    `json:"timestamp,omitempty"`
	Models    []string `json:"models,omitempty"`
}

type HistoryMessage struct {
	ID        string   `json:"id"`
	Role      string   `json:"role"`
	Content   string   `json:"content"`
	Timestamp int64    `json:"timestamp"`
	Models    []string `json:"models"`
}

type History struct {
	CurrentID string                    `json:"current_id"`
	Messages  map[string]HistoryMessage `json:"messages"`
}

type Chat struct {
	Title    string             `json:"title"`
	Models   []string           `json:"models"`
	Messages []OpenWebUIMessage `json:"messages"`
	History  History            `json:"history"`
}

type OpenWebUIChatCreateRequest struct {
	Chat Chat `json:"chat"`
}

type Meta map[string]interface{}

type OpenWebUIChatRequest struct {
	Model    string             `json:"model"`
	Title    string             `json:"title"`
	System   string             `json:"system"`
	Messages []OpenWebUIMessage `json:"messages"`
}

type BackgroundTasks struct {
	TitleGeneration    bool `json:"title_generation"`
	TagsGeneration     bool `json:"tags_generation"`
	FollowUpGeneration bool `json:"follow_up_generation"`
}

type Features struct {
	CodeInterpreter bool `json:"code_interpreter"`
	WebSearch       bool `json:"web_search"`
	ImageGeneration bool `json:"image_generation"`
	Memory          bool `json:"memory"`
}

type OpenWebUICompletion struct {
	ChatID          string             `json:"chat_id"`
	ID              string             `json:"id"`
	Messages        []OpenWebUIMessage `json:"messages"`
	Model           string             `json:"model"`
	Stream          bool               `json:"stream"`
	Files           []OpenWebUIFile    `json:"files"`
	SessionID       string             `json:"session_id"`
	BackgroundTasks BackgroundTasks    `json:"background_tasks"`
	Features        Features           `json:"features"`
}

type OpenWebUIChatCreateResponse struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Chat      Chat   `json:"chat"`
	UpdatedAt int64  `json:"updated_at"`
	CreatedAt int64  `json:"created_at"`
	ShareID   *int64 `json:"share_id,omitempty"`
	Archived  bool   `json:"archived"`
	Pinned    bool   `json:"pinned"`
	Meta      Meta   `json:"meta"`
	FolderID  *int64 `json:"folder_id,omitempty"`
}

type OpenWebUICompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int              `json:"index"`
		Message      OpenWebUIMessage `json:"message"`
		FinishReason string           `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type OpenWebUIChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index            int `json:"index"`
		OpenWebUIMessage struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type OpenWebUIFileResponse struct {
	ID            string  `json:"id"`
	UserID        string  `json:"user_id"`
	Hash          *string `json:"hash,omitempty"`
	Filename      string  `json:"filename"`
	Data          Data    `json:"data"`
	Meta          Meta    `json:"meta"`
	CreatedAt     int64   `json:"created_at"`
	UpdatedAt     int64   `json:"updated_at"`
	Status        bool    `json:"status"`
	Path          string  `json:"path"`
	AccessControl *string `json:"access_control,omitempty"`
}

func createNewChat(messageText, sender string) string {
	apikey := os.Getenv("OPENWEBUI_API_KEY")
	url := os.Getenv("OPENWEBUI_URL")
	model := os.Getenv("OPENWEBUI_MODEL")
	newUuid := uuid.New()
	currentTime := time.Now().Unix()

	messageRequest := OpenWebUIChatCreateRequest{
		Chat: Chat{
			Title:  "Chat with " + sender,
			Models: []string{model},
			Messages: []OpenWebUIMessage{
				{
					ID:        newUuid.String(),
					Role:      "user",
					Timestamp: currentTime,
					Models:    []string{model},
					Content:   messageText}},
			History: History{
				CurrentID: newUuid.String(),
				Messages: map[string]HistoryMessage{
					newUuid.String(): {
						ID:        newUuid.String(),
						Role:      "user",
						Timestamp: currentTime,
						Models:    []string{model},
						Content:   messageText,
					},
				}},
		},
	}

	messageBody, _ := json.Marshal(messageRequest)
	fmt.Println("Message Body:", string(messageBody))
	req, err := http.NewRequest(
		"POST",
		"http://"+url+"/api/v1/chats/new",
		bytes.NewBuffer(messageBody),
	)
	if err != nil {
		log.Println("Failed to send typing start:", err)
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
	var response OpenWebUIChatCreateResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
	}

	return response.ID
}

func sendToOpenWebUI(messageText string, fileIds []string) string {
	chatid := os.Getenv("OPENWEBUI_CHAT_ID")
	model := os.Getenv("OPENWEBUI_MODEL")
	apikey := os.Getenv("OPENWEBUI_API_KEY")
	url := os.Getenv("OPENWEBUI_URL")
	messages := []OpenWebUIMessage{
		{
			Role:    "user",
			Content: messageText,
		},
	}

	files := []OpenWebUIFile{}
	for _, element := range fileIds {
		files = append(files,
			OpenWebUIFile{
				Type: "file",
				ID:   element,
			})
	}

	backgroundTasks := BackgroundTasks{
		TitleGeneration:    false,
		TagsGeneration:     false,
		FollowUpGeneration: false,
	}

	features := Features{
		CodeInterpreter: false,
		WebSearch:       false,
		ImageGeneration: false,
		Memory:          false,
	}

	messageData := OpenWebUICompletion{
		Model:           model,
		ChatID:          chatid,
		Stream:          false,
		Messages:        messages,
		Files:           files,
		BackgroundTasks: backgroundTasks,
		Features:        features,
	}

	messageBody, _ := json.Marshal(messageData)
	fmt.Println("Message Body:", string(messageBody))
	req, err := http.NewRequest(
		"POST",
		"http://"+url+"/api/chat/completions",
		bytes.NewBuffer(messageBody),
	)

	if err != nil {
		log.Println("Failed to send typing start:", err)
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
	var response OpenWebUICompletionResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
	}

	responseMessage := response.Choices[0].Message.Content
	return responseMessage
}

func sendFileToOpenWebUI(filename string) string {
	apikey := os.Getenv("OPENWEBUI_API_KEY")
	url := os.Getenv("OPENWEBUI_URL")

	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		log.Fatalf("file does not exist: %s", filename)
	}

	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", filepath.Base(filename))
	io.Copy(part, file)
	writer.Close()

	req, _ := http.NewRequest("POST", "http://"+url+"/api/v1/files/?process=true", body)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+apikey)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	// contentType = writer.FormDataContentType()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	fmt.Println("Response status:", resp.Status)
	respBody, err := io.ReadAll(resp.Body)
	var response OpenWebUIFileResponse
	err = json.Unmarshal(respBody, &response)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
	}

	return response.ID
}

func getOpenWebUIResponse(messageText string, attachments []Attachment) string {
	if len(attachments) > 0 {
		fmt.Println("Files: ", attachments)
		fileIds := uploadFiles(attachments)
		return sendToOpenWebUI(messageText, fileIds)
	}
	return sendToOpenWebUI(messageText, nil)
}

func uploadFiles(attachments []Attachment) []string {
	openWebUIFileIds := []string{}
	log.Println("Attachments: ", attachments)
	for _, element := range attachments {
		contentType := getSignalAttachment(element.ID)
		fileID := sendFileToOpenWebUI(element.ID)
		log.Printf("Uploaded %s of content-type %s.", fileID, contentType)
		openWebUIFileIds = append(openWebUIFileIds, fileID)
	}
	return openWebUIFileIds
}
