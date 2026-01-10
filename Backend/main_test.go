package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCORS(t *testing.T) {
	req, _ := http.NewRequest("OPTIONS", "/anything", nil)
	rr := httptest.NewRecorder()

	handler := corsMiddleware(func(w http.ResponseWriter, r *http.Request) {})
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("CORS OPTIONS harusnya balik 200, tapi dapet %v", rr.Code)
	}
	if rr.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("Header CORS Origin tidak ditemukan")
	}
}

func TestLoginHandler_InvalidBody(t *testing.T) {
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer([]byte("{invalid json}")))
	rr := httptest.NewRecorder()

	loginHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Harusnya bad request, dapet %v", rr.Code)
	}
}

func TestValidateToken_Missing(t *testing.T) {
	req, _ := http.NewRequest("GET", "/protected", nil)
	_, err := validateToken(req)

	if err == nil {
		t.Error("Harusnya error karena token kosong")
	}
}

func TestRegisterHandler_NoAuth(t *testing.T) {
	req, _ := http.NewRequest("POST", "/register", nil)
	rr := httptest.NewRecorder()

	registerHandler(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Harusnya unauthorized, dapet %v", rr.Code)
	}
}

func TestPasswordHashing(t *testing.T) {
	password := "rahasia123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Gagal hashing: %v", err)
	}

	if !CheckPasswordHash(password, hash) {
		t.Error("Password dan Hash harusnya cocok!")
	}

	if CheckPasswordHash("salah_password", hash) {
		t.Error("Harusnya gagal kalau password salah")
	}
}

func TestGetUsersHandler_NoAuth(t *testing.T) {
	req, _ := http.NewRequest("GET", "/users", nil)
	rr := httptest.NewRecorder()

	getUsersHandler(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("Harusnya forbidden, dapet %v", rr.Code)
	}
}

func TestDeleteUserHandler_NoID(t *testing.T) {
	req, _ := http.NewRequest("DELETE", "/users", nil)
	rr := httptest.NewRecorder()

	deleteUserHandler(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Harusnya unauthorized, dapet %v", rr.Code)
	}
}

func TestUpdateUserHandler_NoAuth(t *testing.T) {
	req, _ := http.NewRequest("PATCH", "/users?id=1", bytes.NewBuffer([]byte(`{"nama":"baru"}`)))
	rr := httptest.NewRecorder()

	updateUserHandler(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Harusnya unauthorized, dapet %v", rr.Code)
	}
}

func TestDashboardHandler_NoAuth(t *testing.T) {
	req, _ := http.NewRequest("GET", "/dashboard", nil)
	rr := httptest.NewRecorder()
	dashboardHandler(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Harusnya 401, dapet %v", rr.Code)
	}
}

func TestPelangganHandler_WithQuery(t *testing.T) {
	req, _ := http.NewRequest("GET", "/pelanggan?search=agus&page=1&limit=5", nil)
	rr := httptest.NewRecorder()
	pelangganHandler(rr, req)
}

func TestLayananHandler_CheckRole(t *testing.T) {
	req, _ := http.NewRequest("POST", "/layanan", bytes.NewBuffer([]byte(`{"nama":"Cuci Kilat"}`)))
	rr := httptest.NewRecorder()
	layananHandler(rr, req)
}

func TestLaporanHandler_Types(t *testing.T) {
	req1, _ := http.NewRequest("GET", "/laporan?type=riwayat", nil)
	laporanHandler(httptest.NewRecorder(), req1)

	req2, _ := http.NewRequest("GET", "/laporan?type=keuangan", nil)
	laporanHandler(httptest.NewRecorder(), req2)
}

func TestLaporanExportHandler_Periode(t *testing.T) {
	req1, _ := http.NewRequest("GET", "/export?periode=harian&date=2025-12-19", nil)
	laporanExportHandler(httptest.NewRecorder(), req1)

	req2, _ := http.NewRequest("GET", "/export?periode=bulanan&year=2025&month=12", nil)
	laporanExportHandler(httptest.NewRecorder(), req2)

	req3, _ := http.NewRequest("GET", "/export?periode=tahunan&year=2025", nil)
	laporanExportHandler(httptest.NewRecorder(), req3)
}

func TestGopayHandler(t *testing.T) {
	payload := `{"notification_text": "Anda menerima pembayaran Rp 50.000 dari Customer"}`
	req, _ := http.NewRequest("POST", "/gopay-webhook", strings.NewReader(payload))
	rr := httptest.NewRecorder()

	GopayHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Harusnya OK, dapet %v", rr.Code)
	}

	select {
	case pesan := <-pesanChan:
		if !strings.Contains(pesan, "Rp 50000") {
			t.Errorf("Format pesan salah: %s", pesan)
		}
	default:
		t.Error("Pesan tidak masuk ke channel!")
	}
}

func TestStreamHandler_ContextCancel(t *testing.T) {
	req, _ := http.NewRequest("GET", "/stream", nil)
	ctx, cancel := context.WithCancel(req.Context())
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	go StreamHandler(rr, req)

	cancel()

	t.Log("SSE Handler berhasil diuji untuk case disconnect")
}
