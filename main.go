package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	_ "github.com/serpapi/google-search-results-golang"
	"google.golang.org/api/option"
)

var model *genai.GenerativeModel

var ctx context.Context

var searchApiKey string

func rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello")
}

type Source struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type Article struct {
	Source      Source    `json:"source"`
	Author      string    `json:"author"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Url         string    `json:"url"`
	UrlToImage  string    `json:"urlToImage"`
	PublishedAt time.Time `json:"publishedAt"`
	Content     string    `json:"content"`
}

type NewsResponse struct {
	Status       string    `json:"status"`
	TotalResults int       `json:"totalResults"`
	Articles     []Article `json:"articles"`
}

func newsHandler(w http.ResponseWriter, r *http.Request) {
	topic := r.URL.Query().Get("topic")
	resp, err := getResults(topic)
	if err != nil {
		panic(err)
	}
	summary, err := getSummary(topic, resp)
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(w, fmt.Sprintf("%#v\n", summary))
}

func getSummary(topic string, newsArr *NewsResponse) (string, error) {
	prompt := fmt.Sprintf("You are an conversational AI model designed to create summaries of news given to you on a specific topic. Do NOT use any markdown formatting. Current news about %s\n", topic)
	for i, news := range newsArr.Articles {
		prompt += fmt.Sprintf("Article %d:\nTitle: %s\n-by %s\nDescription: %s\nShortened Content: %s\n",
			i+1,
			news.Title,
			news.Author,
			news.Description,
			news.Content)
	}
	prompt += "User: Summarize the given news\nResponse: "
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", err
	}
	for _, candidate := range resp.Candidates {
		for _, parts := range candidate.Content.Parts {
			fmt.Println(parts)
		}
	}
	return fmt.Sprint(resp.Candidates[0].Content.Parts[0]), nil

}

func getResults(q string) (*NewsResponse, error) {
	from := time.Now().Add(-time.Hour * 24 * 7)
	fmt.Println(from.String())
	reqUrl := fmt.Sprintf("https://newsapi.org/v2/everything?apiKey=%s&q=%s&from=%s&pageSize=%s", searchApiKey, q, from.Format("2006-01-02T15:04:05-0700"), "10")

	fmt.Println(reqUrl)
	req, err := http.NewRequestWithContext(ctx, "GET", reqUrl, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(resp.Body)
	newsResp := &NewsResponse{}
	err = decoder.Decode(newsResp)
	if err != nil {
		return nil, err
	}
	return newsResp, nil

}

func respToArr(resp string) ([]string, error) {
	fmt.Println(resp)
	resp, ok := strings.CutPrefix(resp, "[")
	if !ok {
		return nil, errors.New("Invalid array")
	}
	fmt.Println(resp)
	resp, ok = strings.CutSuffix(resp, "]")
	if !ok {
		return nil, errors.New("Invalid array")
	}
	fmt.Println(resp)
	respSplit := strings.Split(resp, ",")
	for i, s := range respSplit {
		respSplit[i] = strings.Trim(s, "\" ")
	}
	return respSplit, nil
}

func main() {
	ctx = context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("API_KEY")))
	if err != nil {
		panic(err)
	}
	defer client.Close()
	searchApiKey = os.Getenv("SEARCH_API_KEY")
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/news", newsHandler)
	model = client.GenerativeModel("gemini-pro")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
