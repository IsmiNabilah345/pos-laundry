package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/xuri/excelize/v2"
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

type Transaksi struct {
	Kode             string
	CreatedAt        time.Time
	PelangganNama    string
	LayananNama      string
	Berat            float64
	HargaSatuan      float64
	Total            float64
	MetodePembayaran string
	Status           string
}

func laporanHandler(w http.ResponseWriter, r *http.Request) {
	//enableCORS(w)
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

	transaksiURL := BaseURL + "/rest/v1/transaksi?status=eq.selesai&select=*"
	reqTrans, _ := http.NewRequest("GET", transaksiURL, nil)
	reqTrans.Header.Set("Authorization", auth)
	reqTrans.Header.Set("apikey", APIKey)
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
		t := Transaksi{
			Kode:             fmt.Sprint(item["kode"]),
			PelangganNama:    fmt.Sprint(item["pelanggan_nama"]),
			LayananNama:      fmt.Sprint(item["layanan_nama"]),
			MetodePembayaran: fmt.Sprint(item["metode_pembayaran"]),
			Status:           fmt.Sprint(item["status"]),
		}
		if created, ok := item["created_at"].(string); ok {
			if parsed, err := time.Parse(time.RFC3339, created); err == nil {
				t.CreatedAt = parsed
			}
		}
		if total, ok := item["total"].(float64); ok {
			t.Total = total
		}
		if berat, ok := item["berat"].(float64); ok {
			t.Berat = berat
		}
		if t.Berat > 0 {
			t.HargaSatuan = t.Total / t.Berat
			if math.IsNaN(t.HargaSatuan) || math.IsInf(t.HargaSatuan, 0) {
				t.HargaSatuan = 0
			}
		}

		trx = append(trx, t)
	}

	// Buat Excel ke buffer
	f := excelize.NewFile()
	sheet := "Laporan"
	f.NewSheet(sheet)
	index, err := f.GetSheetIndex(sheet)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	f.SetActiveSheet(index)
	if len(trx) == 0 {
		w.WriteHeader(http.StatusNoContent)
		w.Write([]byte("Tidak ada data transaksi untuk diekspor"))
		return
	}

	headers := []string{"Kode", "Tanggal", "Pelanggan", "Layanan", "Berat", "Harga Satuan", "Total", "Metode Pembayaran", "Status"}
	for i, h := range headers {
		col := string(rune('A' + i))
		f.SetCellValue(sheet, col+"1", h)
	}

	for idx, t := range trx {
		row := strconv.Itoa(idx + 2)
		f.SetCellValue(sheet, "A"+row, t.Kode)
		f.SetCellValue(sheet, "B"+row, t.CreatedAt.Format("02 Jan 2006 15:04"))
		f.SetCellValue(sheet, "C"+row, t.PelangganNama)
		f.SetCellValue(sheet, "D"+row, t.LayananNama)
		f.SetCellValue(sheet, "E"+row, t.Berat)
		f.SetCellValue(sheet, "F"+row, t.HargaSatuan)
		f.SetCellValue(sheet, "G"+row, t.Total)
		f.SetCellValue(sheet, "H"+row, t.MetodePembayaran)
		f.SetCellValue(sheet, "I"+row, t.Status)
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=Laporan_Keuangan.xlsx")
	w.Write(buf.Bytes())
}

func laporanJSONHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	auth := r.Header.Get("Authorization")
	if auth == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "missing token",
		})
		return
	}

	// ambil data transaksi (ringkas)
	transaksiURL := BaseURL + "/rest/v1/transaksi?status=eq.selesai&select=*"
	req, _ := http.NewRequest("GET", transaksiURL, nil)
	req.Header.Set("Authorization", auth)
	req.Header.Set("apikey", APIKey)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	var list []map[string]any
	json.NewDecoder(res.Body).Decode(&list)

	var total float64
	for _, t := range list {
		if v, ok := t["total"].(float64); ok {
			total += v
		}
	}

	json.NewEncoder(w).Encode(map[string]any{
		"data":        list,
		"total_omset": total,
	})
}

func main() {
	http.HandleFunc("/login", corsMiddleware(loginHandler))
	http.HandleFunc("/dashboard", corsMiddleware(dashboardHandler))
	http.HandleFunc("/pelanggan", corsMiddleware(pelangganHandler))
	http.HandleFunc("/transaksi", corsMiddleware(transaksiHandler))
	http.HandleFunc("/layanan", corsMiddleware(layananHandler))
	http.HandleFunc("/laporan", corsMiddleware(laporanHandler))
	http.HandleFunc("/laporan-json", corsMiddleware(laporanJSONHandler))

	fmt.Println("Server jalan di http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
