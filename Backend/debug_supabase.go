package main

import (
	"fmt"
	"io"
	"net/http"
)

func DebugSupabase() {
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
