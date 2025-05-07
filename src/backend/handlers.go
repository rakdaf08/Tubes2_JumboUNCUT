// src/backend/handlers.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	// sync tidak perlu di sini
)

// MultiSearchResponse struct (tetap sama)
type MultiSearchResponse struct {
	SearchTarget   string              `json:"searchTarget"`
	Algorithm      string              `json:"algorithm"`
	Mode           string              `json:"mode"`
	MaxRecipes     int                 `json:"maxRecipes,omitempty"`
	PathFound      bool                `json:"pathFound"`
	Path           []Recipe            `json:"path,omitempty"` // For shortest mode
	Paths          [][]Recipe          `json:"paths,omitempty"` // For multiple mode
	ImageURLs      map[string]string   `json:"imageURLs,omitempty"`
	NodesVisited   int                 `json:"nodesVisited"`
	DurationMillis int64               `json:"durationMillis"`
	Error          string              `json:"error,omitempty"`
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
	if algo == "" { algo = "bfs" }
	if mode == "" { mode = "shortest" }

	// 2. Validasi Input Dasar
	if targetElement == "" { http.Error(w, "Parameter 'target' diperlukan", http.StatusBadRequest); return }
	// Gunakan IsElementExists dari data.go atau file lain yang sesuai
	if !IsElementExists(targetElement) { http.Error(w, fmt.Sprintf("Elemen target '%s' tidak valid", targetElement), http.StatusBadRequest); return }
	if algo != "bfs" && algo != "dfs" { http.Error(w, "Parameter 'algo' harus 'bfs' atau 'dfs'", http.StatusBadRequest); return }
	if mode != "shortest" && mode != "multiple" { http.Error(w, "Parameter 'mode' harus 'shortest' atau 'multiple'", http.StatusBadRequest); return }

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
		SearchTarget:   targetElement,
		Algorithm:      algo,
		Mode:           mode,
		ImageURLs:      make(map[string]string), // Inisialisasi map gambar
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
			log.Printf("Peringatan: Mode 'shortest' dengan algo 'dfs' belum diimplementasikan. Menjalankan BFS.")
			singlePath, nodesVisited, err = FindPathDFS(targetElement) // Fallback ke BFS
		}
		// Set hasil untuk mode shortest
		response.Path = singlePath
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

	// --- Ambil URL Gambar ---
	if response.PathFound {
        imgMap := GetImageMap() // Pastikan fungsi ini ada dan mengembalikan map[string]string
        elementsInPaths := make(map[string]bool)
        elementsInPaths[targetElement] = true // Selalu tambahkan target

        pathsToProcess := [][]Recipe{}
        if mode == "shortest" {
            if len(singlePath) > 0 {
                pathsToProcess = append(pathsToProcess, singlePath)
            }
        } else {
            pathsToProcess = multiplePaths // Gunakan hasil dari multiplePaths
        }

        // Kumpulkan semua elemen unik dari semua jalur yang relevan
        for _, path := range pathsToProcess {
            for _, step := range path {
                elementsInPaths[step.Ingredient1] = true
                elementsInPaths[step.Ingredient2] = true
                // elementsInPaths[step.Result] = true // Opsional: tambahkan hasil perantara
            }
        }

		// Tambahkan elemen dasar jika relevan
		baseElements := []string{"Air", "Earth", "Fire", "Water"}
		for _, base := range baseElements {
			// Cek jika gambar ada DAN elemen dasar relevan (target atau ada di path)
			if imgUrl, exists := imgMap[base]; exists && imgUrl != "" {
				if (isBaseElement(targetElement) && targetElement == base) || elementsInPaths[base] {
                    // Hanya tambahkan jika belum ada (meskipun seharusnya tidak terjadi untuk base)
                    if _, present := response.ImageURLs[base]; !present {
					    response.ImageURLs[base] = imgUrl
                    }
				}
			}
		}

		// Ambil URL untuk elemen non-dasar yang unik
		for elementName := range elementsInPaths {
            // Hanya tambahkan jika belum ada DAN bukan elemen dasar (sudah ditangani)
            if _, isBase := find(baseElements, elementName); !isBase {
                 if imgUrl, ok := imgMap[elementName]; ok && imgUrl != "" {
                    if _, exists := response.ImageURLs[elementName]; !exists {
                         response.ImageURLs[elementName] = imgUrl
                    }
                 }
            }
		}
	}

	// Encode Response ke JSON dan Kirim
	w.Header().Set("Content-Type", "application/json")
	jsonResponse, jsonErr := json.MarshalIndent(response, "", "  ")
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
// func isBaseElement(name string) bool {
// 	baseElements := []string{"Air", "Earth", "Fire", "Water"}
// 	for _, base := range baseElements {
// 		if name == base {
// 			return true
// 		}
// 	}
// 	return false
// }

// Fungsi reconstructPathRevised (dari bfs.go) diperlukan jika FindPathDFS (single path) masih digunakan
// Pastikan bisa diakses dari package main
// func reconstructPathRevised(parent map[string]Recipe, target string) []Recipe { ... }

