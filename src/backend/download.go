package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type ElementImage struct {
	Name     string `json:"name"`
	ImageURL string `json:"imageURL"`
}

const (
	// Jika image_downloader.go ada di .../src/backend/
	// dan dijalankan dari .../src/backend/, maka path relatifnya adalah:
	jsonFilePath    = "data/element_images_urls.json" // Path ke file JSON, relatif dari CWD
	outputDirData   = "data"                          // Direktori 'data' relatif dari CWD
	outputDirImages = "image"                         // Subdirektori 'image' di dalam 'outputDirData'

	maxConcurrentDownloads = 10
	requestTimeout       = 30 * time.Second
)

// ... (sisa fungsi sanitizeFilename dan getFileExtension tetap sama) ...
func sanitizeFilename(name string) string {
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
		" ", "_",
	)
	return replacer.Replace(name)
}

func getFileExtension(rawURL string, contentType string) string {
	parsedURL, err := url.Parse(rawURL)
	if err == nil {
		ext := filepath.Ext(parsedURL.Path)
		if ext != "" && len(ext) > 1 {
			cleanExt := strings.Split(strings.ToLower(ext), "?")[0]
			if len(cleanExt) > 1 && len(cleanExt) <= 5 {
				return cleanExt
			}
		}
	}
	if contentType != "" {
		switch strings.ToLower(strings.Split(contentType, ";")[0]) {
		case "image/jpeg":
			return ".jpg"
		case "image/png":
			return ".png"
		case "image/gif":
			return ".gif"
		case "image/svg+xml":
			return ".svg"
		case "image/webp":
			return ".webp"
		case "image/avif":
			return ".avif"
		}
	}
	log.Printf("Tidak dapat menentukan ekstensi untuk URL: %s, Content-Type: '%s'. Menggunakan default .png\n", rawURL, contentType)
	return ".png"
}


func downloadFile(element ElementImage, fullOutputDir string, client *http.Client, wg *sync.WaitGroup, sem chan struct{}) {
	defer wg.Done()
	defer func() { <-sem }()

	fmt.Printf("Mencoba mengunduh: '%s' dari %s\n", element.Name, element.ImageURL)

	if element.ImageURL == "" {
		log.Printf("URL gambar kosong untuk elemen: '%s'\n", element.Name)
		return
	}

	req, err := http.NewRequest("GET", element.ImageURL, nil)
	if err != nil {
		log.Printf("Gagal membuat request untuk '%s' (%s): %v\n", element.Name, element.ImageURL, err)
		return
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Gagal mengunduh '%s' (%s): %v\n", element.Name, element.ImageURL, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Gagal mengunduh '%s' (%s): status code %d\n", element.Name, element.ImageURL, resp.StatusCode)
		return
	}

	baseFileName := element.Name // Menggunakan nama elemen langsung
	contentType := resp.Header.Get("Content-Type")
	fileExt := getFileExtension(element.ImageURL, contentType)
	finalFileName := baseFileName + fileExt
	filePath := filepath.Join(fullOutputDir, finalFileName)

	out, err := os.Create(filePath)
	if err != nil {
		log.Printf("Gagal membuat file '%s' untuk '%s': %v\n", filePath, element.Name, err)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Printf("Gagal menyimpan gambar '%s' ke '%s': %v\n", element.Name, filePath, err)
		os.Remove(filePath)
		return
	}

	fmt.Printf("Berhasil mengunduh '%s' -> '%s'\n", element.Name, filePath)
}

func maxConcurrentDownloads() {
	log.Println("Mulai proses pengunduhan gambar...")

	// 1. Dapatkan CWD untuk referensi logging
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Gagal mendapatkan current working directory: %v", err)
	}
	fmt.Printf("Current Working Directory: %s\n", cwd)

	// Path ke file JSON, sekarang relatif terhadap CWD
	// Jika CWD adalah .../src/backend/, maka jsonFilePath harus "data/element_images_urls.json"
	// Jika CWD adalah .../src/, maka jsonFilePath harus "backend/data/element_images_urls.json"
	// Sesuaikan `jsonFilePath` di atas jika CWD Anda berbeda.

	// Menggunakan path yang sudah didefinisikan di konstanta
	absJsonFilePath, err := filepath.Abs(jsonFilePath)
	if err != nil {
		log.Fatalf("Gagal mendapatkan path absolut untuk file JSON '%s' (dari CWD '%s'): %v\n", jsonFilePath, cwd, err)
	}
	fmt.Printf("Mencoba membaca file JSON dari: %s\n", absJsonFilePath)

	jsonData, err := os.ReadFile(absJsonFilePath)
	if err != nil {
		log.Fatalf("Gagal membaca file JSON '%s': %v\n", absJsonFilePath, err)
	}

	var elements []ElementImage
	err = json.Unmarshal(jsonData, &elements)
	if err != nil {
		log.Fatalf("Gagal unmarshal data JSON: %v\n", err)
	}

	if len(elements) == 0 {
		log.Println("Tidak ada elemen ditemukan dalam file JSON.")
		return
	}
	fmt.Printf("Ditemukan %d elemen gambar untuk diunduh.\n", len(elements))

	// 2. Buat direktori output: CWD/data/image/
	// Path output akhir relatif terhadap CWD
	finalOutputDirRel := filepath.Join(outputDirData, outputDirImages)
	absFinalOutputDir, err := filepath.Abs(finalOutputDirRel)
	if err != nil {
		log.Fatalf("Gagal mendapatkan path absolut untuk direktori output '%s' (dari CWD '%s'): %v\n", finalOutputDirRel, cwd, err)
	}

	if _, err := os.Stat(absFinalOutputDir); os.IsNotExist(err) {
		errDir := os.MkdirAll(absFinalOutputDir, 0755)
		if errDir != nil {
			log.Fatalf("Gagal membuat direktori output '%s': %v\n", absFinalOutputDir, errDir)
		}
		fmt.Printf("Direktori output '%s' berhasil dibuat.\n", absFinalOutputDir)
	} else {
		fmt.Printf("Direktori output '%s' sudah ada.\n", absFinalOutputDir)
	}

	// 3. Setup untuk unduhan konkurensi
	var wg sync.WaitGroup
	sem := make(chan struct{}, maxConcurrentDownloads)

	customTransport := &http.Transport{
		TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
	}

	client := &http.Client{
		Timeout:   requestTimeout,
		Transport: customTransport,
	}

	// 4. Iterasi dan unduh setiap gambar
	for _, el := range elements {
		if el.ImageURL == "" {
			log.Printf("Melewati elemen '%s' karena URL gambar kosong.\n", el.Name)
			continue
		}
		wg.Add(1)
		sem <- struct{}{}
		go downloadFile(el, absFinalOutputDir, client, &wg, sem)
	}

	wg.Wait()
	close(sem)

	log.Println("Proses pengunduhan gambar selesai.")
	fmt.Printf("Gambar seharusnya telah diunduh ke: %s\n", absFinalOutputDir)
}