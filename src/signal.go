package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type Attachment struct {
	ContentType     string  `json:"contentType"`
	Filename        *string `json:"filename,omitempty"`
	ID              string  `json:"id"`
	Size            int64   `json:"size"`
	Width           int     `json:"width"`
	Height          int     `json:"height"`
	Caption         *string `json:"caption,omitempty"`
	UploadTimestamp int64   `json:"uploadTimestamp"`
}

type DataMessage struct {
	Timestamp          int64        `json:"timestamp"`
	Message            string       `json:"message"`
	ExpiresInSeconds   int          `json:"expiresInSeconds"`
	IsExpirationUpdate bool         `json:"isExpirationUpdate"`
	ViewOnce           bool         `json:"viewOnce"`
	Attachments        []Attachment `json:"attachments"`
}

type Envelope struct {
	Source          string       `json:"source"`
	SourceNumber    string       `json:"sourceNumber"`
	SourceUuid      string       `json:"sourceUuid"`
	SourceName      string       `json:"sourceName"`
	SourceDevice    int          `json:"sourceDevice"`
	Timestamp       int64        `json:"timestamp"`
	ServerReceived  int64        `json:"serverReceivedTimestamp"`
	ServerDelivered int64        `json:"serverDeliveredTimestamp"`
	DataMessage     *DataMessage `json:"dataMessage"`
}

type LinkPreview struct {
	Base64Thumbnail string `json:"base64_thumbnail"`
	Description     string `json:"description"`
	Title           string `json:"title"`
	URL             string `json:"url"`
}

type Mention struct {
	Author string `json:"author"`
	Length int    `json:"length"`
	Start  int    `json:"start"`
}

type QuoteMention struct {
	Author string `json:"author"`
	Length int    `json:"length"`
	Start  int    `json:"start"`
}

type SignalMessageResponse struct {
	Base64Attachments []string `json:"base64_attachments,omitempty"`
	EditTimestamp     int64    `json:"edit_timestamp,omitempty"`
	//LinkPreview       LinkPreview    `json:"link_preview,omitempty"`
	Mentions       []Mention      `json:"mentions,omitempty"`
	Message        string         `json:"message"`
	NotifySelf     bool           `json:"notify_self,omitempty"`
	Number         string         `json:"number"`
	QuoteAuthor    string         `json:"quote_author,omitempty"`
	QuoteMentions  []QuoteMention `json:"quote_mentions,omitempty"`
	QuoteMessage   string         `json:"quote_message,omitempty"`
	QuoteTimestamp int64          `json:"quote_timestamp,omitempty"`
	Recipients     []string       `json:"recipients"`
	Sticker        string         `json:"sticker,omitempty"`
	TextMode       string         `json:"text_mode,omitempty"`
	ViewOnce       bool           `json:"view_once,omitempty"`
}

type SignalTypingRequest struct {
	Recipient string `json:"recipient"`
}

type SignalMessage struct {
	Envelope Envelope `json:"envelope"`
	Account  string   `json:"account"`
}

func getSignalAttachment(attachmentId string) string {
	signalUrl := os.Getenv("SIGNAL_URL")
	out, err := os.Create(attachmentId)
	contentType := ""
	if err != nil {
		log.Println("Failed to create attachment file:", err)
	}

	req, err := http.NewRequest(
		"GET",
		"http://"+signalUrl+"/v1/attachments/"+attachmentId,
		nil,
	)
	if err != nil {
		log.Println("Failed to send typing start:", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		fmt.Println("Response:", resp)
	}
	defer resp.Body.Close()

	fmt.Println("Attachment response status:", resp.Status)
	contentType = resp.Header.Get("content-type")
	fmt.Println("Attachment type: ", contentType)
	defer out.Close()
	io.Copy(out, resp.Body)
	if os.Getenv("DEBUG") == "1" {
		fmt.Fprintln(os.Stdout, string(resp.Status))
		bufio.NewWriter(os.Stdout).Flush()
	}

	return contentType
}

func sendTypingIndicator(action string, accountNumber string, sender string) {
	typingData := SignalTypingRequest{
		Recipient: sender,
	}

	typingBody, _ := json.Marshal(typingData)
	signalUrl := os.Getenv("SIGNAL_URL")
	go func() {
		req, err := http.NewRequest(
			action,
			"http://"+signalUrl+"/v1/typing-indicator/"+accountNumber,
			bytes.NewBuffer(typingBody),
		)
		if err != nil {
			log.Println("Failed to send typing start:", err)
		}

		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error sending request:", err)
			fmt.Println("Response:", resp)
			return
		}
		defer resp.Body.Close()

		fmt.Println("Typing response status:", resp.Status)

		if os.Getenv("DEBUG") == "1" {
			fmt.Fprintln(os.Stdout, string(resp.Status))
			bufio.NewWriter(os.Stdout).Flush()
		}
	}()
}

func sendSignalMessage(message string, account string, sender string) {
	signalUrl := os.Getenv("SIGNAL_URL")
	signalMessage := SignalMessageResponse{
		Message:    message,
		Number:     account,
		Recipients: []string{sender},
	}

	messageBody, _ := json.Marshal(signalMessage)
	fmt.Println("Message Body:", string(messageBody))
	req, err := http.NewRequest(
		"POST",
		"http://"+signalUrl+"/v2/send",
		bytes.NewBuffer(messageBody),
	)

	if err != nil {
		log.Println("Failed to send typing start:", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Send message response status:", resp.Status)

	if os.Getenv("DEBUG") == "1" {
		fmt.Fprintln(os.Stdout, string(resp.Status))
		bufio.NewWriter(os.Stdout).Flush()
	}
}
