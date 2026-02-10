package handlers

import (
	"bytes"
	"io"
	"log"
	"mime/multipart"
	"net/http"
)

const colabScanURL = "https://YOUR_NGROK_URL_HERE/scan"

func scanPicture(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// limit upload size (example: 10MB)
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Invalid multipart form", http.StatusBadRequest)
		return
	}

	file, fileHeader, err := r.FormFile("picture")
	if err != nil {
		http.Error(w, "picture is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Prepare multipart request to Colab
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("file", fileHeader.Filename)
	if err != nil {
		http.Error(w, "Failed to create form file", http.StatusInternalServerError)
		return
	}

	if _, err := io.Copy(part, file); err != nil {
		http.Error(w, "Failed to copy file", http.StatusInternalServerError)
		return
	}

	// Optional: forward extra metadata
	_ = writer.WriteField("source", "outpatient-backend")

	writer.Close()

	req, err := http.NewRequest(http.MethodPost, colabScanURL, &body)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("COLAB ERROR:", err)
		http.Error(w, "Failed to reach AI service", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read AI response", http.StatusInternalServerError)
		return
	}

	// proxy response directly
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(respBody)
}
