package server

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)
func ExtractUUIDFromLink(link string) (string, string, error) {
    u, err := url.Parse(link)
    if err != nil {
        return "", "", err
    }

    // Get the last segment of the path (uuid + extension)
    filename := path.Base(u.Path) // e.g. "550e8400-e29b-41d4-a716-446655440000.png"

    // Split into uuid and extension
    parts := strings.SplitN(filename, ".", 2)
    if len(parts) != 2 {
        return "", "", fmt.Errorf("invalid filename format")
    }

    uuid := parts[0]
    ext := parts[1]

    return uuid, ext, nil
}

func MakeImageLink(baseURL, uuid, ext string) string {
    return fmt.Sprintf("%s/uploads/%s.%s", baseURL, uuid, ext)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
    err := r.ParseMultipartForm(10 << 20) // 10 MB max
    if err != nil {
        http.Error(w, "Could not parse form", http.StatusBadRequest)
        return
    }

    file, handler, err := r.FormFile("file")
    if err != nil {
        http.Error(w, "Could not get file", http.StatusBadRequest)
        return
    }
    defer file.Close()

    uuid := r.FormValue("uuid")
    if uuid == "" {
        http.Error(w, "Missing UUID", http.StatusBadRequest)
        return
    }

    savePath := fmt.Sprintf("uploads/%s%s", uuid, filepath.Ext(handler.Filename))
    out, err := os.Create(savePath)
    if err != nil {
		fmt.Println("Error creating file:", savePath, err)
		http.Error(w, "Could not save file", http.StatusInternalServerError)
		return
	}
    defer out.Close()

    //io.Copy(out, file)
	_, err = io.Copy(out, file)
	if err != nil {
		fmt.Println("Error writing file:", err)
		http.Error(w, "Could not write file", http.StatusInternalServerError)
		return
	}

    fmt.Fprintf(w, "File uploaded successfully: %s\n", savePath)
}
