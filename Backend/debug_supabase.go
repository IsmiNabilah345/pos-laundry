package main

import (
	"fmt"
	"io"
	"net/http"
)

const (
	BaseURL = "https://knyteuovymlwcqnywtmm.supabase.co"
	APIKey  = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImtueXRldW92eW1sd2Nxbnl3dG1tIiwicm9sZSI6ImFub24iLCJpYXQiOjE3NjQyNTM2MjUsImV4cCI6MjA3OTgyOTYyNX0.LXMhoqQZbVfTfZjRBbv2LpCMpGu8qR6iD2NAtva1wJY"
)

func main() {
	url := fmt.Sprintf("%s/rest/v1/users?select=*", BaseURL)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("apikey", APIKey)
	req.Header.Set("Authorization", "Bearer "+APIKey)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("Connection Error: %v\n", err)
		return
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	fmt.Printf("Status: %s\n", res.Status)
	fmt.Printf("Body: %s\n", string(body))
}
