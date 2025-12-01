package client

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
)

//GenerateImageLink takes a local filename and returns a unique HTTP link
func GenerateImageLink(baseURL, filename string) (string, string) {
    //generate a random UUID
    id := uuid.New().String()

    //extract the file extension (e.g. .png, .jpg)
    ext := filepath.Ext(filename)

    //construct the link using baseURL + UUID + extension
    return fmt.Sprintf("%s/%s%s", baseURL, id, ext), id
}

//uploadImageToServer sends a file + UUID to the server's /upload endpoint
func UploadImageToServer(serverURL, uuid, filePath string) (string, error) {
    // Open the original file
    file, err := os.Open(filePath)
    if err != nil {
        return "", err
    }
    defer file.Close()

    //decode the image
    img, _, err := image.Decode(file)
    if err != nil {
        return "", fmt.Errorf("decode failed: %w", err)
    }

    //resize to max width 800px
    resized := imaging.Resize(img, 800, 0, imaging.Lanczos)

    //encode to JPEG with quality 70
    buf := &bytes.Buffer{}
    err = jpeg.Encode(buf, resized, &jpeg.Options{Quality: 70})
    if err != nil {
        return "", fmt.Errorf("encode failed: %w", err)
    }

    //prepare multipart form
    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)

    //add UUID field
    err = writer.WriteField("uuid", uuid)
    if err != nil {
        return "", err
    }

    //add file field (always .jpg now)
    part, err := writer.CreateFormFile("file", uuid+".jpg")
    if err != nil {
        return "", err
    }
    _, err = io.Copy(part, buf)
    if err != nil {
        return "", err
    }

    //close writer
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

    //build usable URL
    usableURL := fmt.Sprintf("%s/uploads/%s.jpg", serverURL, uuid)
    return usableURL, nil
}