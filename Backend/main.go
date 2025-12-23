package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	BaseURL    = "https://knyteuovymlwcqnywtmm.supabase.co"
	APIKey     = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImtueXRldW92eW1sd2Nxbnl3dG1tIiwicm9sZSI6ImFub24iLCJpYXQiOjE3NjQyNTM2MjUsImV4cCI6MjA3OTgyOTYyNX0.LXMhoqQZbVfTfZjRBbv2LpCMpGu8qR6iD2NAtva1wJY"
	AdminEmail = "admin@example.com"
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

	fullURL := BaseURL + "/auth/v1/token?grant_type=password"

	client := &http.Client{}
	reqSup, _ := http.NewRequest("POST", fullURL, bytes.NewBuffer(body))
	reqSup.Header.Set("Content-Type", "application/json")
	reqSup.Header.Set("apikey", APIKey)
	reqSup.Header.Set("Authorization", "Bearer "+APIKey)

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

	role := "kasir"
	if userObj, ok := supRes["user"].(map[string]any); ok {
		if email, ok := userObj["email"].(string); ok && email == AdminEmail {
			role = "admin"
		}
	} else if req.Username == AdminEmail {
		role = "admin"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"access_token": token,
		"role":         role,
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

	client := &http.Client{}

	userURL := BaseURL + "/auth/v1/user"
	reqUser, _ := http.NewRequest("GET", userURL, nil)
	reqUser.Header.Set("Authorization", auth)
	reqUser.Header.Set("apikey", APIKey)

	resUser, err := client.Do(reqUser)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "supabase error"})
		return
	}
	defer resUser.Body.Close()

	var userRes map[string]any
	if err := json.NewDecoder(resUser.Body).Decode(&userRes); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "decode error"})
		return
	}

	if resUser.StatusCode != http.StatusOK {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid token"})
		return
	}

	//nanti diubah jadi per 10 transaksi terakhir supaya tidak lemot ketika banyak data
	transaksiURL := BaseURL + "/rest/v1/transaksi?select=*&order=created_at.desc"
	reqTrans, _ := http.NewRequest("GET", transaksiURL, nil)
	reqTrans.Header.Set("Authorization", auth)
	reqTrans.Header.Set("apikey", APIKey)

	resTrans, err := client.Do(reqTrans)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "supabase error"})
		return
	}
	defer resTrans.Body.Close()

	var transaksiList []map[string]any
	if resTrans.StatusCode == http.StatusOK {
		json.NewDecoder(resTrans.Body).Decode(&transaksiList)
	} else {
		transaksiList = []map[string]any{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"ok":        true,
		"user":      userRes["email"],
		"transaksi": transaksiList,
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

	// forward query string juga, pencarian
	fullURL := BaseURL + "/rest/v1/pelanggan"
	if r.URL.RawQuery != "" {
		fullURL += "?" + r.URL.RawQuery
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)

	client := &http.Client{}
	reqSup, _ := http.NewRequest(r.Method, fullURL, buf)
	reqSup.Header.Set("Content-Type", "application/json")
	reqSup.Header.Set("Authorization", auth)
	reqSup.Header.Set("apikey", APIKey)

	resSup, err := client.Do(reqSup)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "supabase error"})
		return
	}
	defer resSup.Body.Close()

	bodyBytes, _ := io.ReadAll(resSup.Body)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resSup.StatusCode)
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

	fullURL := BaseURL + "/rest/v1/transaksi"
	if r.URL.RawQuery != "" {
		fullURL += "?" + r.URL.RawQuery
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)

	client := &http.Client{}
	reqSup, _ := http.NewRequest(r.Method, fullURL, buf)
	reqSup.Header.Set("Content-Type", "application/json")
	reqSup.Header.Set("Authorization", auth)
	reqSup.Header.Set("apikey", APIKey)

	resSup, err := client.Do(reqSup)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "supabase error"})
		return
	}
	defer resSup.Body.Close()

	bodyBytes, _ := io.ReadAll(resSup.Body)
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

	// RBAC: Hanya Admin yang boleh Post/Delete/Patch
	if r.Method == http.MethodPost || r.Method == http.MethodDelete || r.Method == http.MethodPatch || r.Method == http.MethodPut {
		client := &http.Client{}
		userURL := BaseURL + "/auth/v1/user"
		reqUser, _ := http.NewRequest("GET", userURL, nil)
		reqUser.Header.Set("Authorization", auth)
		reqUser.Header.Set("apikey", APIKey)

		resUser, err := client.Do(reqUser)
		if err == nil {
			defer resUser.Body.Close()
			var userRes map[string]any
			if json.NewDecoder(resUser.Body).Decode(&userRes) == nil {
				if email, ok := userRes["email"].(string); !ok || email != AdminEmail {
					w.WriteHeader(http.StatusForbidden)
					json.NewEncoder(w).Encode(map[string]string{"error": "Akses ditolak: Hanya admin yang boleh mengubah layanan"})
					return
				}
			}
		}
	}

	fullURL := BaseURL + "/rest/v1/layanan"
	if r.URL.RawQuery != "" {
		fullURL += "?" + r.URL.RawQuery
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)

	client := &http.Client{}
	reqSup, _ := http.NewRequest(r.Method, fullURL, buf)
	reqSup.Header.Set("Content-Type", "application/json")
	reqSup.Header.Set("Authorization", auth)
	reqSup.Header.Set("apikey", APIKey)

	resSup, err := client.Do(reqSup)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "supabase error"})
		return
	}
	defer resSup.Body.Close()

	bodyBytes, _ := io.ReadAll(resSup.Body)
	w.Header().Set("Content-Type", "application/json")
	w.Write(bodyBytes)
}

func laporanHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	auth := r.Header.Get("Authorization")
	if auth == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "missing token"})
		return
	}

	// RBAC Check for View Reports (Optional? User asked to create reports)
	// We will enforce Admin only for reports as implied by "Admin vs Kasir" request
	client := &http.Client{}
	userURL := BaseURL + "/auth/v1/user"
	reqUser, _ := http.NewRequest("GET", userURL, nil)
	reqUser.Header.Set("Authorization", auth)
	reqUser.Header.Set("apikey", APIKey)
	resUser, err := client.Do(reqUser)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer resUser.Body.Close()
	var userRes map[string]any
	json.NewDecoder(resUser.Body).Decode(&userRes)
	if email, ok := userRes["email"].(string); !ok || email != AdminEmail {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	// Fetch ALL transactions
	// Warning: Fetching *all* might be heavy, but for POS MVP it's fine.
	transaksiURL := BaseURL + "/rest/v1/transaksi?select=*&status=eq.selesai"
	reqTrans, _ := http.NewRequest("GET", transaksiURL, nil)
	reqTrans.Header.Set("Authorization", auth)
	reqTrans.Header.Set("apikey", APIKey)

	resTrans, err := client.Do(reqTrans)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "supabase error"})
		return
	}
	defer resTrans.Body.Close()

	// Let's use map to be safe

	// Let's use map to be safe
	var rawList []map[string]any
	if err := json.NewDecoder(resTrans.Body).Decode(&rawList); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "decode error"})
		return
	}

	totalOmset := 0.0
	count := 0

	for _, item := range rawList {
		// Parse Total
		var val float64
		switch v := item["total"].(type) {
		case string:
			fmt.Sscanf(v, "%f", &val)
		case float64:
			val = v
		}
		totalOmset += val
		count++
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"total_omset":     totalOmset,
		"total_transaksi": count,
		"data":            rawList, // Send raw data to frontend for grouping
	})
}


func main() {
	http.HandleFunc("/login", corsMiddleware(loginHandler))
	http.HandleFunc("/dashboard", corsMiddleware(dashboardHandler))
	http.HandleFunc("/pelanggan", corsMiddleware(pelangganHandler))
	http.HandleFunc("/transaksi", corsMiddleware(transaksiHandler))
	http.HandleFunc("/layanan", corsMiddleware(layananHandler))
	http.HandleFunc("/laporan", corsMiddleware(laporanHandler))
	fmt.Println("Server jalan di http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
