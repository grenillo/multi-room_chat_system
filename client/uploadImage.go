package client

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// GenerateImageLink takes a local filename and returns a unique HTTP link
// that can later be served by your chat server.
func GenerateImageLink(baseURL, filename string) (string, string) {
    // Generate a random UUID
    id := uuid.New().String()

    // Extract the file extension (e.g. .png, .jpg)
    ext := filepath.Ext(filename)

    // Construct the link using baseURL + UUID + extension
    return fmt.Sprintf("%s/%s%s", baseURL, id, ext), id
}

// UploadImageToServer sends a file + UUID to the server's /upload endpoint
func UploadImageToServer(serverURL, uuid, filePath string) (string, error) {
    // Open the file
    file, err := os.Open(filePath)
    if err != nil {
        return "", err
    }
    defer file.Close()

    // Create a buffer and multipart writer
    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)
    // Add UUID field
    err = writer.WriteField("uuid", uuid)
    if err != nil {
        return "", err
    }

    // Add file field
    part, err := writer.CreateFormFile("file", filepath.Base(filePath))
    if err != nil {
        return "", err
    }
    _, err = io.Copy(part, file)
    if err != nil {
        return "", err
    }

    // Close writer to finalize the form
    err = writer.Close()
    if err != nil {
        return "", err
    }
    // Send POST request
    req, err := http.NewRequest("POST", serverURL+"/upload", body)
    if err != nil {
        return "", err
    }
    req.Header.Set("Content-Type", writer.FormDataContentType())

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("upload failed: %s", resp.Status)
    }

    fmt.Println("Upload successful")
   // Build the usable URL string for others
    ext := filepath.Ext(filePath)
    usableURL := fmt.Sprintf("%s/uploads/%s%s", serverURL, uuid, ext)
    return usableURL, nil
}