package main

import (
	"fmt"
	"net/http"
	"log"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
	_ "github.com/serpapi/google-search-results-golang"
	"context"
	"os"
	"strings"
	"errors"
	"time"
)

var model *genai.GenerativeModel

var ctx context.Context

var searchApiKey string

func rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello")
}

func newsHandler(w http.ResponseWriter, r *http.Request) {
	topic := r.URL.Query().Get("topic")
	prompt := fmt.Sprintf("User: Provide the google search terms for getting information about %s as of %s. Your response is an array of the terms in the format [\"term\"] of lenght 5. You Don't have to perform the search, just give what needs to be searched.\nResponse: ", topic, time.Now().Format("2 January 2006"))
	fmt.Println(prompt)
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		panic(err)
	}
	respSplit, err := respToArr(fmt.Sprint(resp.Candidates[0].Content.Parts[0]))
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(w, fmt.Sprint(respSplit))
}

func getResults(q string) []string {
	reqUrl := fmt.Sprintf("https://newsapi.org/v2/everything")
	req, err := http.NewRequestWithContext(ctx, "GET", "https://newsapi.org/", nil)	
	if err != nil {
		panic(err)
	}
	from := time.Now().Add(-time.Hour * 24 * 7)
	fmt.Println(from.String())
	req.URL.Query().Add("apiKey", searchApiKey)
	req.URL.Query().Add("q", q)
	req.URL.Query().Add("from", from.Format(time.RFC3339))
}

func respToArr(resp string) ([]string, error) {
	fmt.Println(resp)
	resp, ok := strings.CutPrefix(resp, "[");	
	if !ok {
		return nil, errors.New("Invalid array")
	}
	fmt.Println(resp)
	resp, ok = strings.CutSuffix(resp, "]");	
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
	http.HandleFunc("/", rootHandler);
	http.HandleFunc("/news", newsHandler);
	model = client.GenerativeModel("gemini-pro")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
