// src/backend/handlers.go
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url" // Pastikan package ini sudah di-import
	"strconv"
	"strings"
	"time"
	// sync tidak perlu di sini
)

// MultiSearchResponse struct (tetap sama)
type MultiSearchResponse struct {
	SearchTarget   string            `json:"searchTarget"`
	Algorithm      string            `json:"algorithm"`
	Mode           string            `json:"mode"`
	MaxRecipes     int               `json:"maxRecipes,omitempty"`
	PathFound      bool              `json:"pathFound"`
	Path           []Recipe          `json:"path,omitempty"`  // For shortest mode
	Paths          [][]Recipe        `json:"paths,omitempty"` // For multiple mode
	ImageURLs      map[string]string `json:"imageURLs,omitempty"`
	NodesVisited   int               `json:"nodesVisited"`
	DurationMillis int64             `json:"durationMillis"`
	Error          string            `json:"error,omitempty"`
}

// imageHandler function (tetap sama seperti sebelumnya)
func imageHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers agar frontend bisa mengakses
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Hanya izinkan metode GET
	if r.Method != http.MethodGet {
		http.Error(w, "Metode tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	// Ambil nama elemen dari query parameter 'elementName'
	queryParams := r.URL.Query()
	elementName := queryParams.Get("elementName")

	if elementName == "" {
		http.Error(w, "Parameter 'elementName' diperlukan", http.StatusBadRequest)
		return
	}

	log.Printf("Menerima permintaan gambar untuk elemen: %s\n", elementName)

	// Dapatkan map URL gambar dari data yang sudah dimuat
	imageMap := GetImageMap() // Pastikan fungsi GetImageMap() tersedia dari data.go

	// Cari URL gambar asli untuk elemen ini
	originalImageURL, found := imageMap[elementName]
	if !found || originalImageURL == "" {
		log.Printf("URL gambar tidak ditemukan untuk elemen: %s\n", elementName)
		http.Error(w, "URL gambar tidak ditemukan", http.StatusNotFound)
		return
	}

	log.Printf("Mengambil gambar dari URL: %s\n", originalImageURL)

	// Lakukan permintaan HTTP GET ke URL gambar asli DARI BACKEND
	client := http.Client{
		Timeout: 10 * time.Second, // Tambahkan timeout untuk request eksternal
	}
	req, err := http.NewRequest("GET", originalImageURL, nil)
	if err != nil {
		log.Printf("Gagal membuat request ke URL gambar eksternal %s: %v\n", originalImageURL, err)
		http.Error(w, "Gagal mengambil gambar", http.StatusInternalServerError)
		return
	}
	// Opsional: Coba tambahkan User-Agent agar terlihat seperti browser sungguhan
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; MyLittleAlchemyApp/1.0)")
	// Opsional: Coba tambahkan Referer jika diperlukan oleh server sumber
	// req.Header.Set("Referer", "http://littlealchemy2.com/")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Gagal melakukan permintaan GET ke URL gambar eksternal %s: %v\n", originalImageURL, err)
		http.Error(w, "Gagal mengambil gambar", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close() // Pastikan body respons ditutup

	// Periksa status code dari respons server sumber gambar
	if resp.StatusCode != http.StatusOK {
		log.Printf("Server sumber gambar mengembalikan status non-OK untuk %s: %d\n", originalImageURL, resp.StatusCode)
		http.Error(w, fmt.Sprintf("Gagal mengambil gambar dari sumber (%d)", resp.StatusCode), resp.StatusCode)
		return
	}

	// Salin header Content-Type dari respons server sumber gambar ke respons backend kita
	contentType := resp.Header.Get("Content-Type")
	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	} else {
		// Jika Content-Type tidak ada, coba tebak atau default ke SVG
		if strings.HasSuffix(strings.ToLower(originalImageURL), ".svg") {
			w.Header().Set("Content-Type", "image/svg+xml")
		} else if strings.HasSuffix(strings.ToLower(originalImageURL), ".png") {
			w.Header().Set("Content-Type", "image/png")
		} // Tambahkan tipe lain jika perlu
	}
	// Salin header lain yang relevan jika perlu (misal Cache-Control)
	// w.Header().Set("Cache-Control", resp.Header.Get("Cache-Control"))

	// Salin body respons (data gambar) dari server sumber ke respons backend
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("Gagal menyalin body respons gambar dari %s: %v\n", originalImageURL, err)
		return // Keluar saja setelah logging
	}

	log.Printf("Gambar untuk elemen %s berhasil dilayani.\n", elementName)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Hanya izinkan metode GET
	if r.Method != http.MethodGet {
		http.Error(w, "Metode tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	// 1. Ambil Query Parameters
	targetElement := strings.TrimSpace(r.URL.Query().Get("target"))
	algo := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("algo")))
	mode := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("mode")))
	maxRecipesStr := r.URL.Query().Get("max")

	// Default values
	if algo == "" {
		algo = "bfs"
	}
	if mode == "" {
		mode = "shortest"
	}

	// 2. Validasi Input Dasar
	if targetElement == "" {
		http.Error(w, "Parameter 'target' diperlukan", http.StatusBadRequest)
		return
	}
	// Gunakan IsElementExists dari data.go atau file lain yang sesuai
	if !IsElementExists(targetElement) {
		http.Error(w, fmt.Sprintf("Elemen target '%s' tidak valid", targetElement), http.StatusBadRequest)
		return
	}
	if algo != "bfs" && algo != "dfs" {
		http.Error(w, "Parameter 'algo' harus 'bfs' atau 'dfs'", http.StatusBadRequest)
		return
	}
	if mode != "shortest" && mode != "multiple" {
		http.Error(w, "Parameter 'mode' harus 'shortest' atau 'multiple'", http.StatusBadRequest)
		return
	}

	// 3. Proses parameter 'max' jika mode 'multiple'
	maxRecipes := 1
	if mode == "multiple" {
		if maxRecipesStr != "" {
			var convErr error
			maxRecipes, convErr = strconv.Atoi(maxRecipesStr)
			if convErr != nil || maxRecipes <= 0 {
				http.Error(w, "Parameter 'max' harus berupa angka positif untuk mode 'multiple'", http.StatusBadRequest)
				return
			}
		} else {
			http.Error(w, "Parameter 'max' diperlukan untuk mode 'multiple'", http.StatusBadRequest)
			return
		}
	}

	// 4. Panggil Fungsi Algoritma & Ukur Waktu
	startTime := time.Now()

	var singlePath []Recipe
	var multiplePaths [][]Recipe
	var nodesVisited int
	var err error
	var pathFound bool

	log.Printf("Memulai pencarian: Target=%s, Algo=%s, Mode=%s, Max=%d\n", targetElement, algo, mode, maxRecipes)

	// --- Struktur Response Awal ---
	response := MultiSearchResponse{
		SearchTarget: targetElement,
		Algorithm:    algo,
		Mode:         mode,
		// ImageURLs akan diinisialisasi nanti setelah path ditemukan
	}
	if mode == "multiple" {
		response.MaxRecipes = maxRecipes // Set max recipes jika mode multiple
	}

	// --- Logika Pemilihan Algoritma ---
	if mode == "shortest" {
		if algo == "bfs" {
			singlePath, nodesVisited, err = FindPathBFS(targetElement)
		} else { // algo == "dfs"
			// TODO: Implementasi DFS Shortest (IDDFS) jika diperlukan
			log.Printf("Peringatan: Mode 'shortest' dengan algo 'dfs' belum diimplementasikan. Menjalankan DFS Single Path.")
			singlePath, nodesVisited, err = FindPathDFS(targetElement) // Menggunakan DFS Single Path
		}
		// Set hasil untuk mode shortest
		response.Path = singlePath
		// pathFound true jika tidak ada error DAN (path tidak kosong ATAU target adalah elemen dasar)
		pathFound = err == nil && (len(singlePath) > 0 || (len(singlePath) == 0 && isBaseElement(targetElement)))

	} else { // mode == "multiple"
		if algo == "bfs" {
			multiplePaths, nodesVisited, err = FindMultiplePathsBFS(targetElement, maxRecipes)
		} else { // algo == "dfs"
			// --- PANGGIL FUNGSI DFS MULTIPLE BARU ---
			log.Printf("Menjalankan DFS Multiple untuk target: %s, max: %d", targetElement, maxRecipes)
			multiplePaths, nodesVisited, err = FindMultiplePathsDFS(targetElement, maxRecipes) // Panggil fungsi baru
			// -----------------------------------------
		}
		// Set hasil untuk mode multiple
		response.Paths = multiplePaths
		// pathFound true jika tidak ada error DAN (paths tidak kosong ATAU target adalah elemen dasar)
		pathFound = err == nil && (len(multiplePaths) > 0 || (len(multiplePaths) == 0 && isBaseElement(targetElement)))
	}

	duration := time.Since(startTime)
	log.Printf("Pencarian selesai: Durasi=%v, Nodes=%d, Ditemukan=%t, Error=%v\n", duration, nodesVisited, pathFound, err)

	// --- Isi sisa response ---
	response.PathFound = pathFound
	response.NodesVisited = nodesVisited // Ingat: -1 untuk DFS Multiple saat ini
	response.DurationMillis = duration.Milliseconds()

	if err != nil {
		response.Error = err.Error()
	}

	// --- Ambil URL Gambar untuk SEMUA elemen yang relevan ---
	// Bagian ini diperbaiki untuk memastikan URL proxy yang benar digunakan
	if response.PathFound {
		imgMap := GetImageMap() // Pastikan fungsi ini ada dan mengembalikan map[string]string
		elementsInPaths := make(map[string]bool)

		// Kumpulkan semua elemen unik dari semua jalur resep yang berhasil ditemukan
		pathsToProcess := [][]Recipe{}
		if response.Mode == "shortest" && response.Path != nil {
			if len(response.Path) > 0 { // Hanya tambahkan path jika tidak kosong
				pathsToProcess = append(pathsToProcess, response.Path)
			}
		} else if response.Mode == "multiple" && response.Paths != nil {
			if len(response.Paths) > 0 { // Hanya tambahkan paths jika tidak kosong
				pathsToProcess = response.Paths
			}
		}

		// Kumpulkan semua elemen unik dari semua jalur yang berhasil ditemukan
		for _, path := range pathsToProcess {
			for _, step := range path {
				elementsInPaths[step.Ingredient1] = true
				elementsInPaths[step.Ingredient2] = true
				elementsInPaths[step.Result] = true // Tambahkan juga elemen hasil di setiap langkah
			}
		}
		// Tambahkan target elemen ke elementsInPaths jika belum ada
		elementsInPaths[response.SearchTarget] = true

		response.ImageURLs = make(map[string]string) // Inisialisasi map gambar di sini setelah tahu elemen mana yang relevan

		// Iterasi SEMUA elemen relevan (dari paths + target) dan tambahkan URL proxy mereka
		for elementName := range elementsInPaths {
			// Cek apakah elemen ini punya gambar di map asli
			if imgUrl, ok := imgMap[elementName]; ok && imgUrl != "" {
				// BUAT URL YANG MENGARAH ke endpoint backend proxy /api/image
				// Gunakan url.QueryEscape untuk memastikan nama elemen aman dalam query string URL
				proxyUrl := fmt.Sprintf("/api/image?elementName=%s", url.QueryEscape(elementName))

				// Simpan URL proxy RELATIF di response.ImageURLs
				// Cek apakah sudah ada (seharusnya tidak ada karena map baru diinisialisasi)
				if _, exists := response.ImageURLs[elementName]; !exists {
					response.ImageURLs[elementName] = proxyUrl
				}
			} else {
				// Opsional: Jika elemen tidak punya URL gambar di map asli, bisa dicatat atau diabaikan
				// fmt.Printf("Info: URL gambar tidak ditemukan di map asli untuk elemen '%s'.\n", elementName)
			}
		}
	}
	// END --- Ambil URL Gambar ---

	// Encode Response ke JSON dan Kirim
	w.Header().Set("Content-Type", "application/json")
	jsonResponse, jsonErr := json.MarshalIndent(response, "", "  ") // Perbaiki indentasi
	if jsonErr != nil {
		log.Printf("Error saat marshal JSON response: %v", jsonErr)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	_, writeErr := w.Write(jsonResponse)
	if writeErr != nil {
		log.Printf("Error saat menulis JSON response: %v", writeErr)
	}
}

// Helper function find (jika belum ada)
func find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

// Pastikan fungsi isBaseElement ada dan bisa diakses
// (Bisa didefinisikan di sini, di dfs.go, atau di data.go)
// func isBaseElement(name string) bool { ... }

// Pastikan fungsi reconstructPathRevised (dari bfs.go) diperlukan jika FindPathDFS (single path) masih digunakan
// Pastikan bisa diakses dari package main
// func reconstructPathRevised(parent map[string]Recipe, target string) []Recipe { ... }
