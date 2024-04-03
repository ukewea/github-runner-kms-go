package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type GitHubTokenResponse struct {
	Token string `json:"token"`
}

var orgPatMap map[string]string

func loadOrgPatMap() {
	jsonData, err := ioutil.ReadFile("org-pat-map.json")
	if err != nil {
		fmt.Printf("Error reading JSON file: %v\n", err)
	}

	err = json.Unmarshal(jsonData, &orgPatMap)
	if err != nil {
		fmt.Printf("Error parsing JSON file: %v\n", err)
	}
}

// Function to get the proxy URL from environment variables
func getProxyURL() (*url.URL, error) {
	proxyURLStr := os.Getenv("HTTPS_PROXY")
	if proxyURLStr == "" {
		proxyURLStr = os.Getenv("https_proxy") // Check for lowercase if uppercase not found
	}
	if proxyURLStr == "" {
		return nil, nil // No proxy set
	}
	return url.Parse(proxyURLStr)
}

func getGithubToken(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 2 {
		http.Error(w, "Invalid request path", http.StatusBadRequest)
		return
	}
	orgName, tokenType := parts[0], parts[1]

	var tokenURLPath string
	if tokenType == "registration-token" {
		tokenURLPath = "/actions/runners/registration-token"
	} else if tokenType == "remove-token" {
		tokenURLPath = "/actions/runners/remove-token"
	} else {
		http.Error(w, "Invalid token type", http.StatusBadRequest)
		return
	}

	pat, exists := orgPatMap[orgName]
	if !exists {
		http.Error(w, "PAT not found for organization", http.StatusNotFound)
		return
	}

	proxyURL, err := getProxyURL()
	if err != nil {
		fmt.Printf("Error parsing proxy URL: %v\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	client := &http.Client{}
	if proxyURL != nil {
		fmt.Printf("Using proxy: %s\n", proxyURL)
		transport := &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
		client.Transport = transport
	}

	reqURL := fmt.Sprintf("https://api.github.com/orgs/%s%s", orgName, tokenURLPath)
	fmt.Printf("Request URL: %s", reqURL)

	req, err := http.NewRequest("POST", reqURL, nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	req.Header.Add("Authorization", "token "+pat)
	response, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var githubResponse GitHubTokenResponse
	err = json.Unmarshal(body, &githubResponse)
	if err != nil {
		fmt.Printf("Error unmarshalling GitHub response: %v\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Send only the token as the response
	w.Write([]byte(githubResponse.Token))
}

func main() {
	proxyURL, err := getProxyURL()
	if err != nil {
		fmt.Printf("Error parsing proxy URL: %v\n", err)
	} else if proxyURL != nil {
		fmt.Printf("Proxy is set to: %s\n", proxyURL)
	} else {
		fmt.Println("Proxy is not set")
	}

	loadOrgPatMap()

	http.HandleFunc("/", getGithubToken)

	fmt.Println("Starting server on port 3000")
	if err := http.ListenAndServe(":3000", nil); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}
