package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Request struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	// MaxTokens define a quantidade de linhas que serão dadas como respostas
	// A partir desse número que é cobrado o chatgpt
	MaxTokens int `json:"max_tokens,omitempty"`
}

type Response struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int      `json:"created"`
	Choices []Choice `json:"choices"`
}

// É possível criar uma struct de uma struct
type Choice struct {
	Index   int `json:"index"`
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
}

func GenerateGPTText(query string) (string, error) {
	req := Request{
		Model: "gpt-3.5-turbo",
		Messages: []Message{
			{
				Role:    "user",
				Content: query,
			},
		},
		MaxTokens: 150,
	}
	reqJson, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	// Montar a request request para o chatGpt
	request, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(reqJson))
	if err != nil {
		return "", err
	}
	// Adicionar os Headers
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer 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")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	respBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	var resp Response
	err = json.Unmarshal(respBody, &resp)
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}

// O twilio manda o request em form
func parseBase64RequestData(r string) (string, error) {
	dataBytes, err := base64.StdEncoding.DecodeString(r)
	if err != nil {
		return "", err
	}
	data, err := url.ParseQuery(string(dataBytes))
	if data.Has("Body") {
		return data.Get("Body"), nil
	}
	return "", errors.New("Body not found")
}
func process(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	result, err := parseBase64RequestData(request.Body)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       err.Error(),
		}, nil
	}
	text, err := GenerateGPTText(result)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       err.Error(),
		}, nil
	}
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       text,
	}, nil
}
func main() {
	lambda.Start(process)
}
