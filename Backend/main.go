package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/xuri/excelize/v2"
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

	if !CheckPasswordHash(req.Password, user.Password) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Password salah",
		})
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

func registerHandler(w http.ResponseWriter, r *http.Request) {

	claims, err := validateToken(r)
	if err != nil {
		fmt.Println("Gagal Verifikasi Karena:", err)

		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	if claims["role"] != "admin" {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Hanya admin yang bisa menambah kasir"})
		return
	}

	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Nama     string `json:"nama"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	hashed, err := HashPassword(input.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	userData := map[string]any{
		"username": input.Username,
		"password": hashed,
		"nama":     input.Nama,
		"role":     "kasir",
	}

	body, _ := json.Marshal(userData)

	client := &http.Client{}
	url := BaseURL + "/rest/v1/users"
	reqSup, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	reqSup.Header.Set("apikey", APIKey)
	reqSup.Header.Set("Authorization", "Bearer "+APIKey)
	reqSup.Header.Set("Content-Type", "application/json")

	resSup, err := client.Do(reqSup)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer resSup.Body.Close()

	if resSup.StatusCode >= 400 {
		w.WriteHeader(resSup.StatusCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Kasir berhasil didaftarkan"})
}

func getUsersHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := validateToken(r)
	if err != nil || claims["role"] != "admin" {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Forbidden"})
		return
	}

	client := &http.Client{}
	// Coba ganti select=* dulu buat ngetes kalau kolom tertentu yang bikin error
	url := BaseURL + "/rest/v1/users?select=*"

	reqSup, _ := http.NewRequest("GET", url, nil)
	reqSup.Header.Set("apikey", APIKey)
	reqSup.Header.Set("Authorization", "Bearer "+APIKey)

	resSup, err := client.Do(reqSup)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer resSup.Body.Close()

	body, _ := io.ReadAll(resSup.Body)

	// Kirim status code yang asli dari Supabase biar keliatan di browser
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resSup.StatusCode)
	w.Write(body)
}

func deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := validateToken(r)
	if err != nil || claims["role"] != "admin" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	client := &http.Client{}
	reqSup, _ := http.NewRequest("DELETE", BaseURL+"/rest/v1/users?id=eq."+id, nil)
	reqSup.Header.Set("apikey", APIKey)
	reqSup.Header.Set("Authorization", "Bearer "+APIKey)

	resSup, _ := client.Do(reqSup)
	w.WriteHeader(resSup.StatusCode)
}

func updateUserHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := validateToken(r)
	if err != nil || claims["role"] != "admin" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	id := r.URL.Query().Get("id")
	var input map[string]interface{}
	json.NewDecoder(r.Body).Decode(&input)

	body, _ := json.Marshal(input)
	client := &http.Client{}
	reqSup, _ := http.NewRequest("PATCH", BaseURL+"/rest/v1/users?id=eq."+id, bytes.NewBuffer(body))
	reqSup.Header.Set("apikey", APIKey)
	reqSup.Header.Set("Authorization", "Bearer "+APIKey)
	reqSup.Header.Set("Content-Type", "application/json")

	resSup, _ := client.Do(reqSup)
	w.WriteHeader(resSup.StatusCode)
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
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

	query := r.URL.Query()
	searchTerm := query.Get("search")

	bodyBytes, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	filterQuery := ""
	if searchTerm != "" {
		filterQuery = fmt.Sprintf("&or=(nama.ilike.*%s*,telepon.ilike.*%s*)", searchTerm, searchTerm)
	}

	fullURL := BaseURL + "/rest/v1/pelanggan?select=*"
	if r.Method == http.MethodGet {
		fullURL += filterQuery
	}

	page, _ := strconv.Atoi(query.Get("page"))
	limit, _ := strconv.Atoi(query.Get("limit"))

	query.Del("page")
	query.Del("limit")
	query.Del("search")

	if len(query) > 0 {
		fullURL += "&" + query.Encode()
	}

	reqSup, _ := http.NewRequest(r.Method, fullURL, bytes.NewBuffer(bodyBytes))

	reqSup.Header.Set("Content-Type", "application/json")
	reqSup.Header.Set("apikey", APIKey)
	reqSup.Header.Set("Authorization", "Bearer "+APIKey)

	if r.Method == http.MethodGet && page > 0 {
		if limit < 1 {
			limit = 10
		}
		from := (page - 1) * limit
		to := from + limit - 1
		reqSup.Header.Set("Range", fmt.Sprintf("%d-%d", from, to))
	}

	client := &http.Client{}
	resSup, err := client.Do(reqSup)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer resSup.Body.Close()

	resBody, _ := io.ReadAll(resSup.Body)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resSup.StatusCode)
	w.Write(resBody)
}

func transaksiHandler(w http.ResponseWriter, r *http.Request) {
	_, err := validateToken(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	query := r.URL.Query()
	page, _ := strconv.Atoi(query.Get("page"))
	limit, _ := strconv.Atoi(query.Get("limit"))
	searchTerm := query.Get("search")

	bodyBytes, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	fullURL := BaseURL + "/rest/v1/transaksi"
	params := url.Values{}
	params.Add("select", "*")
	params.Add("order", "created_at.desc")

	if r.Method == http.MethodGet && searchTerm != "" {
		params.Add("kode", "ilike.*"+searchTerm+"*")
	}

	query.Del("page")
	query.Del("limit")
	query.Del("search")
	query.Del("order")
	for k, v := range query {
		params.Add(k, v[0])
	}
	fullURL += "?" + params.Encode()

	client := &http.Client{}
	reqSup, _ := http.NewRequest(r.Method, fullURL, bytes.NewBuffer(bodyBytes))
	reqSup.Header.Set("Content-Type", "application/json")
	reqSup.Header.Set("apikey", APIKey)
	reqSup.Header.Set("Authorization", "Bearer "+APIKey)

	if r.Method == http.MethodGet && page > 0 {
		if limit < 1 {
			limit = 10
		}
		from := (page - 1) * limit
		to := from + limit - 1
		reqSup.Header.Set("Range", fmt.Sprintf("%d-%d", from, to))
	}

	resSup, err := client.Do(reqSup)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer resSup.Body.Close()

	resBody, _ := io.ReadAll(resSup.Body)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resSup.StatusCode)
	w.Write(resBody)
}

func layananHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := validateToken(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	method := r.Method
	if method == http.MethodPost || method == http.MethodDelete || method == http.MethodPatch || method == http.MethodPut {
		role, _ := claims["role"].(string)
		if role != "admin" {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{"error": "Akses ditolak: Hanya admin"})
			return
		}
	}

	query := r.URL.Query()
	page, _ := strconv.Atoi(query.Get("page"))
	limit, _ := strconv.Atoi(query.Get("limit"))
	searchTerm := query.Get("search")

	bodyBytes, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	fullURL := BaseURL + "/rest/v1/layanan"
	params := url.Values{}
	params.Add("select", "*")

	if method == http.MethodGet && searchTerm != "" {
		params.Add("nama", "ilike.*"+searchTerm+"*")
	}

	query.Del("page")
	query.Del("limit")
	query.Del("search")
	for k, v := range query {
		params.Add(k, v[0])
	}

	fullURL += "?" + params.Encode()

	client := &http.Client{}
	reqSup, _ := http.NewRequest(method, fullURL, bytes.NewBuffer(bodyBytes))
	reqSup.Header.Set("Content-Type", "application/json")
	reqSup.Header.Set("apikey", APIKey)
	reqSup.Header.Set("Authorization", "Bearer "+APIKey)

	if method == http.MethodGet && page > 0 {
		if limit < 1 {
			limit = 10
		}
		from := (page - 1) * limit
		to := from + limit - 1
		reqSup.Header.Set("Range", fmt.Sprintf("%d-%d", from, to))
	}

	resSup, err := client.Do(reqSup)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer resSup.Body.Close()

	resBody, _ := io.ReadAll(resSup.Body)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resSup.StatusCode)
	w.Write(resBody)
}

type Transaksi struct {
	Kode             string     `json:"kode"`
	CreatedAt        time.Time  `json:"created_at"`
	SelesaiAt        *time.Time `json:"selesai_at"`
	PelangganNama    string     `json:"pelanggan_nama"`
	LayananNama      string     `json:"layanan_nama"`
	Berat            float64    `json:"berat"`
	HargaSatuan      float64    `json:"harga_satuan"`
	Total            float64    `json:"total"`
	MetodePembayaran string     `json:"metode_pembayaran"`
	Status           string     `json:"status"`
}

func laporanHandler(w http.ResponseWriter, r *http.Request) {
	_, err := validateToken(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	tipe := r.URL.Query().Get("type")
	client := &http.Client{}

	if tipe == "riwayat" {
		transaksiURL := BaseURL +
			"/rest/v1/transaksi_wib" +
			"?status=eq.selesai" +
			"&select=kode,created_at,selesai_at,berat,total,metode_pembayaran,status,pelanggan(nama),layanan(nama)" +
			"&order=selesai_at.desc"

		req, _ := http.NewRequest("GET", transaksiURL, nil)
		req.Header.Set("apikey", APIKey)
		req.Header.Set("Authorization", "Bearer "+APIKey)

		res, err := client.Do(req)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer res.Body.Close()

		var raw []map[string]any
		json.NewDecoder(res.Body).Decode(&raw)

		trx := []Transaksi{}
		for _, item := range raw {
			t := Transaksi{}

			t.Kode = fmt.Sprint(item["kode"])
			t.Status = fmt.Sprint(item["status"])
			t.MetodePembayaran = fmt.Sprint(item["metode_pembayaran"])

			if p, ok := item["pelanggan"].(map[string]any); ok {
				t.PelangganNama = fmt.Sprint(p["nama"])
			}
			if l, ok := item["layanan"].(map[string]any); ok {
				t.LayananNama = fmt.Sprint(l["nama"])
			}

			if s, ok := item["selesai_at"].(string); ok && s != "" {
				if tm, err := time.Parse(time.RFC3339, s); err == nil {
					t.SelesaiAt = &tm
				}
			}

			if total, ok := item["total"].(float64); ok {
				t.Total = total
			}

			trx = append(trx, t)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"data": trx})
		return
	}

	if tipe == "keuangan" {
		loc, _ := time.LoadLocation("Asia/Jakarta")
		today := time.Now().In(loc).Format("2006-01-02")

		getSum := func(url string) float64 {
			req, _ := http.NewRequest("GET", url, nil)
			req.Header.Set("apikey", APIKey)
			req.Header.Set("Authorization", "Bearer "+APIKey)

			res, err := client.Do(req)
			if err != nil {
				return 0
			}
			defer res.Body.Close()

			var list []map[string]any
			json.NewDecoder(res.Body).Decode(&list)

			var sum float64
			for _, i := range list {
				if t, ok := i["total"].(float64); ok {
					sum += t
				}
			}
			return sum
		}

		// TOTAL OMSET (SEMUA)
		totalURL := BaseURL +
			"/rest/v1/transaksi_wib" +
			"?status=eq.selesai&select=total"

		// OMSET HARIAN (WIB FIX)
		dailyURL := BaseURL +
			"/rest/v1/transaksi_wib" +
			"?status=eq.selesai" +
			"&selesai_tanggal_wib=eq." + today +
			"&select=total"

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"total_omset":  getSum(totalURL),
			"omset_harian": getSum(dailyURL),
		})
		return
	}

	w.WriteHeader(http.StatusBadRequest)
}

func laporanExportHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := validateToken(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	role, _ := claims["role"].(string)
	if role != "admin" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	tipe := r.URL.Query().Get("type")
	periode := r.URL.Query().Get("periode")

	loc, _ := time.LoadLocation("Asia/Jakarta")
	var start, end time.Time

	switch periode {
	case "harian":
		valDate := r.URL.Query().Get("date")
		tgl, _ := time.ParseInLocation("2006-01-02", valDate, loc)
		start = time.Date(tgl.Year(), tgl.Month(), tgl.Day(), 0, 0, 0, 0, loc)
		end = time.Date(tgl.Year(), tgl.Month(), tgl.Day(), 23, 59, 59, 0, loc)

	case "bulanan":
		y, _ := strconv.Atoi(r.URL.Query().Get("year"))
		m, _ := strconv.Atoi(r.URL.Query().Get("month"))
		start = time.Date(y, time.Month(m), 1, 0, 0, 0, 0, loc)
		end = start.AddDate(0, 1, -1)
		end = time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 0, loc)

	case "tahunan":
		y, _ := strconv.Atoi(r.URL.Query().Get("year"))
		start = time.Date(y, 1, 1, 0, 0, 0, 0, loc)
		end = time.Date(y, 12, 31, 23, 59, 59, 0, loc)

	default:
		now := time.Now().In(loc)
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
		end = start.AddDate(0, 0, 1)
	}

	filter := fmt.Sprintf("selesai_tanggal_wib=gte.%s&selesai_tanggal_wib=lte.%s",
		start.Format("2006-01-02"),
		end.Format("2006-01-02"))

	var transaksiURL string
	if tipe == "riwayat" {
		transaksiURL = BaseURL + "/rest/v1/transaksi_wib?status=eq.selesai&select=kode,selesai_tanggal_wib,berat,total,metode_pembayaran,status,pelanggan(nama),layanan(nama)"
	} else {
		transaksiURL = BaseURL + "/rest/v1/transaksi_wib?status=eq.selesai&select=selesai_tanggal_wib,total"
	}

	transaksiURL += "&" + filter

	client := &http.Client{}
	req, _ := http.NewRequest("GET", transaksiURL, nil)
	req.Header.Set("apikey", APIKey)
	req.Header.Set("Authorization", "Bearer "+APIKey)

	res, err := client.Do(req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	var raw []map[string]any
	json.NewDecoder(res.Body).Decode(&raw)

	var trx []Transaksi
	var totalOmsetPeriode float64

	for _, item := range raw {
		t := Transaksi{}
		t.Kode = fmt.Sprint(item["kode"])
		t.Status = fmt.Sprint(item["status"])
		t.MetodePembayaran = fmt.Sprint(item["metode_pembayaran"])

		if p, ok := item["pelanggan"].(map[string]any); ok {
			t.PelangganNama = fmt.Sprint(p["nama"])
		}
		if l, ok := item["layanan"].(map[string]any); ok {
			t.LayananNama = fmt.Sprint(l["nama"])
		}

		if s, ok := item["selesai_tanggal_wib"].(string); ok && s != "" {
			if tm, err := time.Parse(time.RFC3339, s); err == nil {
				t.SelesaiAt = &tm
			} else if tm, err := time.Parse("2006-01-02", s); err == nil {
				t.SelesaiAt = &tm
			}
		}

		if total, ok := item["total"].(float64); ok {
			t.Total = total
		} else if totalStr, ok := item["total"].(string); ok {
			t.Total, _ = strconv.ParseFloat(totalStr, 64)
		}

		if berat, ok := item["berat"].(float64); ok {
			t.Berat = berat
		}
		if t.Berat > 0 {
			t.HargaSatuan = t.Total / t.Berat
		}

		totalOmsetPeriode += t.Total
		trx = append(trx, t)
	}

	f := excelize.NewFile()
	sheet := "Laporan"
	f.SetSheetName("Sheet1", sheet)

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"4B0082"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	totalStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"E0E0E0"}, Pattern: 1},
	})

	if tipe == "riwayat" {
		headers := []string{"Kode", "Tanggal Selesai", "Pelanggan", "Layanan", "Berat", "Harga Satuan", "Total", "Metode", "Status"}
		for i, h := range headers {
			col := string(rune('A' + i))
			f.SetCellValue(sheet, col+"1", h)
			f.SetCellStyle(sheet, col+"1", col+"1", headerStyle)
			f.SetColWidth(sheet, col, col, 18)
		}

		for i, t := range trx {
			row := strconv.Itoa(i + 2)
			f.SetCellValue(sheet, "A"+row, t.Kode)
			if t.SelesaiAt != nil {
				f.SetCellValue(sheet, "B"+row, t.SelesaiAt.In(loc).Format("02-01-2006"))
			}
			f.SetCellValue(sheet, "C"+row, t.PelangganNama)
			f.SetCellValue(sheet, "D"+row, t.LayananNama)
			f.SetCellValue(sheet, "E"+row, t.Berat)
			f.SetCellValue(sheet, "F"+row, t.HargaSatuan)
			f.SetCellValue(sheet, "G"+row, t.Total)
			f.SetCellValue(sheet, "H"+row, t.MetodePembayaran)
			f.SetCellValue(sheet, "I"+row, t.Status)
		}

		lastRow := strconv.Itoa(len(trx) + 2)
		f.SetCellValue(sheet, "F"+lastRow, "TOTAL")
		f.SetCellValue(sheet, "G"+lastRow, totalOmsetPeriode)
		f.SetCellStyle(sheet, "F"+lastRow, "G"+lastRow, totalStyle)

	} else {
		f.SetCellValue(sheet, "A1", "Tanggal Selesai")
		f.SetCellValue(sheet, "B1", "Omset (Rp)")
		f.SetCellStyle(sheet, "A1", "B1", headerStyle)
		f.SetColWidth(sheet, "A", "B", 25)

		for i, t := range trx {
			row := strconv.Itoa(i + 2)
			if t.SelesaiAt != nil {
				f.SetCellValue(sheet, "A"+row, t.SelesaiAt.In(loc).Format("02-01-2006"))
			}
			f.SetCellValue(sheet, "B"+row, t.Total)
		}

		lastRow := strconv.Itoa(len(trx) + 2)
		f.SetCellValue(sheet, "A"+lastRow, "TOTAL OMSET PERIODE INI")
		f.SetCellValue(sheet, "B"+lastRow, totalOmsetPeriode)
		f.SetCellStyle(sheet, "A"+lastRow, "B"+lastRow, totalStyle)
	}

	var buf bytes.Buffer
	f.Write(&buf)

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=Laporan_%s_%s.xlsx", tipe, start.Format("2006-01-02")))
	w.Write(buf.Bytes())
}

func main() {
	http.HandleFunc("/login", corsMiddleware(loginHandler))
	http.HandleFunc("/register", corsMiddleware(registerHandler))
	http.HandleFunc("/users", corsMiddleware(getUsersHandler))
	http.HandleFunc("/users/delete", corsMiddleware(deleteUserHandler))
	http.HandleFunc("/users/update", corsMiddleware(updateUserHandler))
	http.HandleFunc("/dashboard", corsMiddleware(dashboardHandler))
	http.HandleFunc("/pelanggan", corsMiddleware(pelangganHandler))
	http.HandleFunc("/transaksi", corsMiddleware(transaksiHandler))
	http.HandleFunc("/layanan", corsMiddleware(layananHandler))
	http.HandleFunc("/laporan", corsMiddleware(laporanHandler))
	http.HandleFunc("/laporan/export", corsMiddleware(laporanExportHandler))
	http.HandleFunc("/api/webhook-gopay", GopayHandler)
	http.HandleFunc("/api/stream", StreamHandler)

	fmt.Println("Server WellClean Laundry jalan di: http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
