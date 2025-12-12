package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type SupabaseLogin struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")

}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid body"})
		return
	}

	supabaseReq := SupabaseLogin{Email: req.Username, Password: req.Password}
	body, _ := json.Marshal(supabaseReq)

	supabaseURL := "https://knyteuovymlwcqnywtmm.supabase.co/auth/v1/token?grant_type=password"
	supabaseKey := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImtueXRldW92eW1sd2Nxbnl3dG1tIiwicm9sZSI6ImFub24iLCJpYXQiOjE3NjQyNTM2MjUsImV4cCI6MjA3OTgyOTYyNX0.LXMhoqQZbVfTfZjRBbv2LpCMpGu8qR6iD2NAtva1wJY"

	client := &http.Client{}
	reqSup, _ := http.NewRequest("POST", supabaseURL, bytes.NewBuffer(body))
	reqSup.Header.Set("Content-Type", "application/json")
	reqSup.Header.Set("apikey", supabaseKey)
	reqSup.Header.Set("Authorization", "Bearer "+supabaseKey)

	resSup, err := client.Do(reqSup)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "supabase error"})
		return
	}
	defer resSup.Body.Close()

	var supRes map[string]any
	if err := json.NewDecoder(resSup.Body).Decode(&supRes); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "decode error"})
		return
	}

	token, ok := supRes["access_token"].(string)
	if !ok || token == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "login gagal"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"access_token": token,
	})

}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	auth := r.Header.Get("Authorization")
	if auth == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "missing token"})
		return
	}

	supabaseURL := "https://knyteuovymlwcqnywtmm.supabase.co/auth/v1/user"
	supabaseKey := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImtueXRldW92eW1sd2Nxbnl3dG1tIiwicm9sZSI6ImFub24iLCJpYXQiOjE3NjQyNTM2MjUsImV4cCI6MjA3OTgyOTYyNX0.LXMhoqQZbVfTfZjRBbv2LpCMpGu8qR6iD2NAtva1wJY"

	client := &http.Client{}
	reqSup, _ := http.NewRequest("GET", supabaseURL, nil)
	reqSup.Header.Set("Authorization", auth) // token dari frontend
	reqSup.Header.Set("apikey", supabaseKey)

	resSup, err := client.Do(reqSup)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "supabase error"})
		return
	}
	defer resSup.Body.Close()

	var userRes map[string]any
	if err := json.NewDecoder(resSup.Body).Decode(&userRes); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "decode error"})
		return
	}

	fmt.Println("Token dari frontend:", auth)
	fmt.Println("Respon Supabase:", userRes)
	if resSup.StatusCode != http.StatusOK {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid token"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"ok":   true,
		"user": userRes["email"],
		"data": []string{"", ""},
	})
}

func pelangganHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	auth := r.Header.Get("Authorization")
	if auth == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "missing token"})
		return
	}

	// forward query string juga
	supabaseURL := "https://knyteuovymlwcqnywtmm.supabase.co/rest/v1/pelanggan"
	if r.URL.RawQuery != "" {
		supabaseURL += "?" + r.URL.RawQuery
	}

	supabaseKey := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImtueXRldW92eW1sd2Nxbnl3dG1tIiwicm9sZSI6ImFub24iLCJpYXQiOjE3NjQyNTM2MjUsImV4cCI6MjA3OTgyOTYyNX0.LXMhoqQZbVfTfZjRBbv2LpCMpGu8qR6iD2NAtva1wJY"

	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)

	client := &http.Client{}
	reqSup, _ := http.NewRequest(r.Method, supabaseURL, buf)
	reqSup.Header.Set("Content-Type", "application/json")
	reqSup.Header.Set("Authorization", auth)
	reqSup.Header.Set("apikey", supabaseKey)

	resSup, err := client.Do(reqSup)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "supabase error"})
		return
	}
	defer resSup.Body.Close()

	bodyBytes, _ := io.ReadAll(resSup.Body)
	fmt.Println("Status:", resSup.StatusCode)
	fmt.Println("Supabase response:", string(bodyBytes))

	w.Header().Set("Content-Type", "application/json")
	w.Write(bodyBytes)
}

func transaksiHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	auth := r.Header.Get("Authorization")
	if auth == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "missing token"})
		return
	}

	supabaseURL := "https://knyteuovymlwcqnywtmm.supabase.co/rest/v1/transaksi"
	if r.URL.RawQuery != "" {
		supabaseURL += "?" + r.URL.RawQuery
	}

	supabaseKey := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImtueXRldW92eW1sd2Nxbnl3dG1tIiwicm9sZSI6ImFub24iLCJpYXQiOjE3NjQyNTM2MjUsImV4cCI6MjA3OTgyOTYyNX0.LXMhoqQZbVfTfZjRBbv2LpCMpGu8qR6iD2NAtva1wJY"

	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)

	client := &http.Client{}
	reqSup, _ := http.NewRequest(r.Method, supabaseURL, buf)
	reqSup.Header.Set("Content-Type", "application/json")
	reqSup.Header.Set("Authorization", auth)
	reqSup.Header.Set("apikey", supabaseKey)

	resSup, err := client.Do(reqSup)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "supabase error"})
		return
	}
	defer resSup.Body.Close()

	bodyBytes, _ := io.ReadAll(resSup.Body)
	fmt.Println("Status:", resSup.StatusCode)
	fmt.Println("Supabase response:", string(bodyBytes))

	w.Header().Set("Content-Type", "application/json")
	w.Write(bodyBytes)
}

func layananHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	auth := r.Header.Get("Authorization")
	if auth == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "missing token"})
		return
	}

	supabaseURL := "https://knyteuovymlwcqnywtmm.supabase.co/rest/v1/layanan" + r.URL.RawQuery
	if r.URL.RawQuery != "" {
		supabaseURL = "https://knyteuovymlwcqnywtmm.supabase.co/rest/v1/layanan?" + r.URL.RawQuery
	}

	supabaseKey := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImtueXRldW92eW1sd2Nxbnl3dG1tIiwicm9sZSI6ImFub24iLCJpYXQiOjE3NjQyNTM2MjUsImV4cCI6MjA3OTgyOTYyNX0.LXMhoqQZbVfTfZjRBbv2LpCMpGu8qR6iD2NAtva1wJY"

	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)

	client := &http.Client{}
	reqSup, _ := http.NewRequest(r.Method, supabaseURL, buf)
	reqSup.Header.Set("Content-Type", "application/json")
	reqSup.Header.Set("Authorization", auth)
	reqSup.Header.Set("apikey", supabaseKey)

	resSup, err := client.Do(reqSup)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "supabase error"})
		return
	}
	defer resSup.Body.Close()

	bodyBytes, _ := io.ReadAll(resSup.Body)
	fmt.Println("Status:", resSup.StatusCode)
	fmt.Println("Supabase response:", string(bodyBytes))

	w.Header().Set("Content-Type", "application/json")
	w.Write(bodyBytes)
}

func main() {
	http.HandleFunc("/login", corsMiddleware(loginHandler))
	http.HandleFunc("/dashboard", corsMiddleware(dashboardHandler))
	http.HandleFunc("/pelanggan", corsMiddleware(pelangganHandler))
	http.HandleFunc("/transaksi", corsMiddleware(transaksiHandler))
	http.HandleFunc("/layanan", corsMiddleware(layananHandler))
	fmt.Println("Server jalan di http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
