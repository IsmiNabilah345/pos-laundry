package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

var pesanChan = make(chan string, 10)

func GopayHandler(w http.ResponseWriter, r *http.Request) {
	bodyBytes, _ := io.ReadAll(r.Body)

	var incoming struct {
		NotificationText string `json:"notification_text"`
	}
	json.Unmarshal(bodyBytes, &incoming)

	re := regexp.MustCompile(`(\d[\d\.]+)`)
	match := re.FindString(incoming.NotificationText)

	finalText := incoming.NotificationText
	if match != "" {
		angkaBersih := strings.ReplaceAll(match, ".", "")
		finalText = "Pembayaran Rp " + angkaBersih + " Berhasil!"
	}

	pesanChan <- finalText

	fmt.Println("Berhasil kirim ke dashboard:", finalText)
	w.WriteHeader(http.StatusOK)
}

func StreamHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	notify := r.Context().Done()

	for {
		select {
		case <-notify:
			return
		case pesan := <-pesanChan:
			fmt.Fprintf(w, "data: %s\n\n", pesan)

			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}
}
