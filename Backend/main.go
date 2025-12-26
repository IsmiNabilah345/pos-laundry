package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/xuri/excelize/v2"
)

const (
	BaseURL   = "https://ndvwwttqjcbkdrnkbsok.supabase.co"
	APIKey    = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6Im5kdnd3dHRxamNia2Rybmtic29rIiwicm9sZSI6ImFub24iLCJpYXQiOjE3NjU5OTIzMzYsImV4cCI6MjA4MTU2ODMzNn0.LWSuFIeQn7WhS3ncYFDd3BxOLXvgtmKiTMgci9xNuLM"
	JWTSecret = "rahasia-pos-laundry-2025" // Secret Key Backend
)

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type LoginRequest struct {
	Username string `json:"username"`
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

// Helper: Validasi JWT
func validateToken(r *http.Request) (jwt.MapClaims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("missing token")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
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

	// 1. Cari user di tabel public.users
	client := &http.Client{}
	url := fmt.Sprintf("%s/rest/v1/users?username=eq.%s&select=*", BaseURL, req.Username)
	reqSup, _ := http.NewRequest("GET", url, nil)
	reqSup.Header.Set("apikey", APIKey)
	reqSup.Header.Set("Authorization", "Bearer "+APIKey)

	resSup, err := client.Do(reqSup)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Supabase Connection Error: " + err.Error()})
		return
	}
	defer resSup.Body.Close()

	if resSup.StatusCode >= 400 {
		var supError map[string]interface{}
		json.NewDecoder(resSup.Body).Decode(&supError)
		w.WriteHeader(resSup.StatusCode)
		json.NewEncoder(w).Encode(map[string]any{"error": "Supabase Error", "details": supError})
		return
	}

	var users []User
	if err := json.NewDecoder(resSup.Body).Decode(&users); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to decode users: " + err.Error()})
		return
	}

	if len(users) == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "User tidak ditemukan"})
		return
	}

	user := users[0]

	// 2. Cek Password (Plain text)
	if user.Password != req.Password {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Password salah"})
		return
	}

	// 3. Buat JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":      user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(time.Hour * 24).Unix(), // 24 jam
	})

	tokenString, err := token.SignedString([]byte(JWTSecret))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"access_token": tokenString,
		"role":         user.Role,
	})
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := validateToken(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	client := &http.Client{}
	// Pakai Key Anon untuk ambil data transaksi
	transaksiURL := BaseURL + "/rest/v1/transaksi?select=*&order=created_at.desc"
	reqTrans, _ := http.NewRequest("GET", transaksiURL, nil)
	reqTrans.Header.Set("apikey", APIKey)
	reqTrans.Header.Set("Authorization", "Bearer "+APIKey)

	resTrans, err := client.Do(reqTrans)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
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
		"user":      claims["username"],
		"role":      claims["role"],
		"transaksi": transaksiList,
	})
}

func pelangganHandler(w http.ResponseWriter, r *http.Request) {
	_, err := validateToken(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	fullURL := BaseURL + "/rest/v1/pelanggan"
	if r.URL.RawQuery != "" {
		fullURL += "?" + r.URL.RawQuery
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)

	client := &http.Client{}
	reqSup, _ := http.NewRequest(r.Method, fullURL, buf)
	reqSup.Header.Set("Content-Type", "application/json")
	reqSup.Header.Set("apikey", APIKey)
	reqSup.Header.Set("Authorization", "Bearer "+APIKey) // Bypass Auth Supabase

	resSup, err := client.Do(reqSup)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer resSup.Body.Close()

	bodyBytes, _ := io.ReadAll(resSup.Body)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resSup.StatusCode)
	w.Write(bodyBytes)
}

func transaksiHandler(w http.ResponseWriter, r *http.Request) {
	_, err := validateToken(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
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
	reqSup.Header.Set("apikey", APIKey)
	reqSup.Header.Set("Authorization", "Bearer "+APIKey)

	resSup, err := client.Do(reqSup)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer resSup.Body.Close()

	bodyBytes, _ := io.ReadAll(resSup.Body)
	w.Header().Set("Content-Type", "application/json")
	w.Write(bodyBytes)
}

func layananHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := validateToken(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// RBAC: Hanya Admin yang boleh Post/Delete/Patch/Put
	method := r.Method
	if method == http.MethodPost || method == http.MethodDelete || method == http.MethodPatch || method == http.MethodPut {
		role, _ := claims["role"].(string)
		if role != "admin" {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{"error": "Akses ditolak: Hanya admin"})
			return
		}
	}

	fullURL := BaseURL + "/rest/v1/layanan"
	if r.URL.RawQuery != "" {
		fullURL += "?" + r.URL.RawQuery
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)

	client := &http.Client{}
	reqSup, _ := http.NewRequest(method, fullURL, buf)
	reqSup.Header.Set("Content-Type", "application/json")
	reqSup.Header.Set("apikey", APIKey)
	reqSup.Header.Set("Authorization", "Bearer "+APIKey)

	resSup, err := client.Do(reqSup)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer resSup.Body.Close()

	bodyBytes, _ := io.ReadAll(resSup.Body)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resSup.StatusCode)
	w.Write(bodyBytes)
}

type Transaksi struct {
	Kode             string    `json:"kode"`
	CreatedAt        time.Time `json:"created_at"`
	PelangganNama    string    `json:"pelanggan_nama"`
	LayananNama      string    `json:"layanan_nama"`
	Berat            float64   `json:"berat"`
	HargaSatuan      float64   `json:"harga_satuan"`
	Total            float64   `json:"total"`
	MetodePembayaran string    `json:"metode_pembayaran"`
	Status           string    `json:"status"`
}

func laporanHandler(w http.ResponseWriter, r *http.Request) {
	_, err := validateToken(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Validasi Role Admin? (Opsional, tapi biasanya laporan utk admin)
	// claims, _ := validateToken(r)
	// if claims["role"] != "admin" { ... }

	tipe := r.URL.Query().Get("type")
	var transaksiURL string
	switch tipe {
	case "riwayat":
		transaksiURL = BaseURL +
			"/rest/v1/transaksi?status=eq.selesai&select=kode,created_at,berat,total,metode_pembayaran,status,pelanggan(nama),layanan(nama)"
	case "keuangan":
		transaksiURL = BaseURL +
			"/rest/v1/transaksi?status=eq.selesai&select=created_at,total"
	default:
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid laporan type"))
		return
	}

	client := &http.Client{}
	reqTrans, _ := http.NewRequest("GET", transaksiURL, nil)
	reqTrans.Header.Set("apikey", APIKey)
	reqTrans.Header.Set("Authorization", "Bearer "+APIKey)

	resTrans, err := client.Do(reqTrans)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer resTrans.Body.Close()

	var rawList []map[string]any
	if err := json.NewDecoder(resTrans.Body).Decode(&rawList); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var trx []Transaksi
	var totalOmset float64

	for _, item := range rawList {
		t := Transaksi{
			Kode:             fmt.Sprint(item["kode"]),
			MetodePembayaran: fmt.Sprint(item["metode_pembayaran"]),
			Status:           fmt.Sprint(item["status"]),
		}

		if p, ok := item["pelanggan"].(map[string]any); ok {
			t.PelangganNama = fmt.Sprint(p["nama"])
		}
		if l, ok := item["layanan"].(map[string]any); ok {
			t.LayananNama = fmt.Sprint(l["nama"])
		}

		if created, ok := item["created_at"].(string); ok {
			if parsed, err := time.Parse(time.RFC3339, created); err == nil {
				t.CreatedAt = parsed
			}
		}

		switch v := item["total"].(type) {
		case float64:
			t.Total = v
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				t.Total = f
			}
		}
		switch v := item["berat"].(type) {
		case float64:
			t.Berat = v
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				t.Berat = f
			}
		}

		if t.Berat > 0 {
			t.HargaSatuan = t.Total / t.Berat
			if math.IsNaN(t.HargaSatuan) || math.IsInf(t.HargaSatuan, 0) {
				t.HargaSatuan = 0
			}
		}

		totalOmset += t.Total
		trx = append(trx, t)
	}

	w.Header().Set("Content-Type", "application/json")

	if tipe == "keuangan" {
		json.NewEncoder(w).Encode(map[string]any{
			"data":        trx,
			"total_omset": totalOmset,
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"data": trx,
	})
}

func laporanExportHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := validateToken(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// RBAC: Hanya Admin
	role, _ := claims["role"].(string)
	if role != "admin" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	tipe := r.URL.Query().Get("type")
	var transaksiURL string
	switch tipe {
	case "riwayat":
		transaksiURL = BaseURL +
			"/rest/v1/transaksi?status=eq.selesai&select=kode,created_at,berat,total,metode_pembayaran,status,pelanggan(nama),layanan(nama)"
	case "keuangan":
		transaksiURL = BaseURL +
			"/rest/v1/transaksi?status=eq.selesai&select=created_at,total"
	default:
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid laporan type"))
		return
	}

	client := &http.Client{}
	reqTrans, _ := http.NewRequest("GET", transaksiURL, nil)
	reqTrans.Header.Set("apikey", APIKey)
	reqTrans.Header.Set("Authorization", "Bearer "+APIKey)

	resTrans, err := client.Do(reqTrans)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer resTrans.Body.Close()

	var rawList []map[string]any
	if err := json.NewDecoder(resTrans.Body).Decode(&rawList); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var trx []Transaksi
	for _, item := range rawList {
		t := Transaksi{}
		if tipe == "riwayat" {
			t.Kode = fmt.Sprint(item["kode"])
			t.MetodePembayaran = fmt.Sprint(item["metode_pembayaran"])
			t.Status = fmt.Sprint(item["status"])
			if p, ok := item["pelanggan"].(map[string]any); ok {
				t.PelangganNama = fmt.Sprint(p["nama"])
			}
			if l, ok := item["layanan"].(map[string]any); ok {
				t.LayananNama = fmt.Sprint(l["nama"])
			}
		}
		if created, ok := item["created_at"].(string); ok {
			t.CreatedAt, _ = time.Parse(time.RFC3339, created)
		}
		if total, ok := item["total"].(float64); ok {
			t.Total = total
		}
		if berat, ok := item["berat"].(float64); ok {
			t.Berat = berat
		}
		if t.Berat > 0 {
			t.HargaSatuan = t.Total / t.Berat
		}
		trx = append(trx, t)
	}

	f := excelize.NewFile()
	sheet := "Laporan"
	f.SetSheetName("Sheet1", sheet)

	if tipe == "riwayat" {
		headers := []string{"Kode", "Tanggal", "Pelanggan", "Layanan", "Berat", "Harga Satuan", "Total", "Metode", "Status"}
		for i, h := range headers {
			f.SetCellValue(sheet, string(rune('A'+i))+"1", h)
		}
		for i, t := range trx {
			r := strconv.Itoa(i + 2)
			f.SetCellValue(sheet, "A"+r, t.Kode)
			f.SetCellValue(sheet, "B"+r, t.CreatedAt.Format("02-01-2006 15:04"))
			f.SetCellValue(sheet, "C"+r, t.PelangganNama)
			f.SetCellValue(sheet, "D"+r, t.LayananNama)
			f.SetCellValue(sheet, "E"+r, t.Berat)
			f.SetCellValue(sheet, "F"+r, t.HargaSatuan)
			f.SetCellValue(sheet, "G"+r, t.Total)
			f.SetCellValue(sheet, "H"+r, t.MetodePembayaran)
			f.SetCellValue(sheet, "I"+r, t.Status)
		}
	} else if tipe == "keuangan" {
		headers := []string{"Tanggal", "Total"}
		for i, h := range headers {
			f.SetCellValue(sheet, string(rune('A'+i))+"1", h)
		}
		for i, t := range trx {
			r := strconv.Itoa(i + 2)
			f.SetCellValue(sheet, "A"+r, t.CreatedAt.Format("02-01-2006"))
			f.SetCellValue(sheet, "B"+r, t.Total)
		}
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	filename := "Laporan.xlsx"
	if tipe == "riwayat" {
		filename = "Riwayat_Transaksi.xlsx"
	} else if tipe == "keuangan" {
		filename = "Laporan_Keuangan.xlsx"
	}
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Write(buf.Bytes())
}

func main() {
	http.HandleFunc("/login", corsMiddleware(loginHandler))
	http.HandleFunc("/dashboard", corsMiddleware(dashboardHandler))
	http.HandleFunc("/pelanggan", corsMiddleware(pelangganHandler))
	http.HandleFunc("/transaksi", corsMiddleware(transaksiHandler))
	http.HandleFunc("/layanan", corsMiddleware(layananHandler))
	http.HandleFunc("/laporan", corsMiddleware(laporanHandler))
	http.HandleFunc("/laporan/export", corsMiddleware(laporanExportHandler))

	fmt.Println("Server jalan di http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
