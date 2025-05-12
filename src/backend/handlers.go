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
)

// MultiSearchResponse struct (tetap sama)
type MultiSearchResponse struct {
	SearchTarget   string            `json:"searchTarget"`
	Algorithm      string            `json:"algorithm"`
	Mode           string            `json:"mode"`
	MaxRecipes     int               `json:"maxRecipes,omitempty"`
	PathFound      bool              `json:"pathFound"`
	Path           []Recipe          `json:"path,omitempty"`
	Paths          [][]Recipe        `json:"paths,omitempty"`
	ImageURLs      map[string]string `json:"imageURLs,omitempty"`
	NodesVisited   int               `json:"nodesVisited"`
	DurationMillis int64             `json:"durationMillis"`
	Error          string            `json:"error,omitempty"`
}


// searchHandler (bagian ImageURLs tetap sama, karena frontend akan memanggil /api/image)
func searchHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method != http.MethodGet {
		http.Error(w, "Metode tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	targetElement := strings.TrimSpace(r.URL.Query().Get("target"))
	// ... (logika validasi targetElement dan title casing tetap sama) ...
	titleCaseTarget := toTitleCase(targetElement)
	firstCapTarget := ""
	if len(targetElement) > 0 {
		firstCapTarget = strings.ToUpper(string(targetElement[0]))
		if len(targetElement) > 1 {
			firstCapTarget += strings.ToLower(targetElement[1:])
		}
	}
	lowerCaseTarget := strings.ToLower(targetElement)
	upperCaseTarget := strings.ToUpper(targetElement)
	potentialTargets := []string{titleCaseTarget, firstCapTarget, targetElement, lowerCaseTarget, upperCaseTarget}
	validTarget := ""
	for _, potTarget := range potentialTargets {
		if IsElementExists(potTarget) {
			validTarget = potTarget
			break
		}
	}
	if validTarget == "" {
		validTarget = titleCaseTarget
	}
	targetElement = validTarget


	algo := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("algo")))
	mode := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("mode")))
	maxRecipesStr := r.URL.Query().Get("max")

	if algo == "" { algo = "bfs" }
	if mode == "" { mode = "shortest" }

	if targetElement == "" {
		http.Error(w, "Parameter 'target' diperlukan", http.StatusBadRequest)
		return
	}
	if !IsElementExists(targetElement) {
		http.Error(w, fmt.Sprintf("Elemen target '%s' tidak valid atau tidak ditemukan", targetElement), http.StatusBadRequest)
		return
	}
	if algo != "bfs" && algo != "dfs" && algo != "bds" {
		http.Error(w, "Parameter 'algo' harus 'bfs', 'dfs', atau 'bds'", http.StatusBadRequest)
		return
	}
	if mode != "shortest" && mode != "multiple" {
		http.Error(w, "Parameter 'mode' harus 'shortest' atau 'multiple'", http.StatusBadRequest)
		return
	}

	maxRecipes := 1
	if mode == "multiple" {
		if maxRecipesStr != "" {
			var convErr error
			maxRecipes, convErr = strconv.Atoi(maxRecipesStr)
			if convErr != nil || maxRecipes <= 0 {
				http.Error(w, "Parameter 'max' harus berupa angka positif lebih besar dari 0 untuk mode 'multiple'", http.StatusBadRequest)
				return
			}
		} else {
			http.Error(w, "Parameter 'max' diperlukan untuk mode 'multiple'", http.StatusBadRequest)
			return
		}
	}

	startTime := time.Now()
	var singlePath []Recipe
	var multiplePaths [][]Recipe
	var nodesVisited int
	var errSearch error
	var pathFound bool

	log.Printf("Memulai pencarian: Target=%s, Algo=%s, Mode=%s, MaxRecipes=%d\n", targetElement, algo, mode, maxRecipes)
	response := MultiSearchResponse{
		SearchTarget: targetElement,
		Algorithm:    algo,
		Mode:         mode,
	}
	if mode == "multiple" {
		response.MaxRecipes = maxRecipes
	}

	if algo == "bfs" {
		if mode == "shortest" {
			singlePath, nodesVisited, errSearch = FindPathBFS(targetElement)
			response.Path = singlePath
			pathFound = errSearch == nil && (len(singlePath) > 0 || (len(singlePath) == 0 && isBaseElement(targetElement)))
		} else {
			multiplePaths, nodesVisited, errSearch = FindMultiplePathsBFS(targetElement, maxRecipes)
			response.Paths = multiplePaths
			pathFound = errSearch == nil && (len(multiplePaths) > 0 || (len(multiplePaths) == 0 && isBaseElement(targetElement)))
		}
	} else if algo == "dfs" {
        if mode == "shortest" {
            singlePath, nodesVisited, errSearch = FindPathDFS(targetElement)
            response.Path = singlePath
            pathFound = errSearch == nil && (len(singlePath) > 0 || (len(singlePath) == 0 && isBaseElementDFS(targetElement)))
        } else {
            multiplePaths, nodesVisited, errSearch = FindMultiplePathsDFS(targetElement, maxRecipes)
            response.Paths = multiplePaths
            pathFound = errSearch == nil && (len(multiplePaths) > 0 || (len(multiplePaths) == 0 && isBaseElementDFS(targetElement)))
        }
	} else if algo == "bds" {
		if mode == "shortest" {
			singlePath, nodesVisited, errSearch = FindPathBDS(targetElement)
			response.Path = singlePath
			pathFound = errSearch == nil && singlePath != nil && (len(singlePath) > 0 || (len(singlePath) == 0 && isBaseElement(targetElement)))
		} else {
			multiplePaths, nodesVisited, errSearch = FindMultiplePathsBDS(targetElement, maxRecipes)
			response.Paths = multiplePaths
			pathFound = errSearch == nil && multiplePaths != nil && (len(multiplePaths) > 0 || (len(multiplePaths) == 0 && isBaseElement(targetElement)))
		}
	}


	duration := time.Since(startTime)
	log.Printf("Pencarian selesai: Durasi=%v, Nodes Dikeluarkan=%d, Path Ditemukan=%t, Error=%v\n", duration, nodesVisited, pathFound, errSearch)

	response.PathFound = pathFound
	response.NodesVisited = nodesVisited
	response.DurationMillis = duration.Milliseconds()
	if errSearch != nil {
		response.Error = errSearch.Error()
	}

	if response.PathFound {
		elementsInPaths := make(map[string]bool)
		pathsToProcess := [][]Recipe{}
		if response.Mode == "shortest" && response.Path != nil {
			if len(response.Path) > 0 {
				pathsToProcess = append(pathsToProcess, response.Path)
			}
		} else if response.Mode == "multiple" && response.Paths != nil {
			if len(response.Paths) > 0 {
				pathsToProcess = response.Paths
			}
		}
		elementsInPaths[response.SearchTarget] = true

		for _, path := range pathsToProcess {
			for _, step := range path {
				elementsInPaths[step.Ingredient1] = true
				elementsInPaths[step.Ingredient2] = true
				elementsInPaths[step.Result] = true
			}
		}

		response.ImageURLs = make(map[string]string)
		for elementName := range elementsInPaths {
			// URL tetap menunjuk ke endpoint /api/image
			// Backend yang akan menangani penyajian file lokal dari endpoint tersebut
			imagePath := fmt.Sprintf("/image/%s.png", elementName)
        	response.ImageURLs[elementName] = imagePath
		}
	}

	w.Header().Set("Content-Type", "application/json")
	jsonResponse, jsonErr := json.MarshalIndent(response, "", "  ")
	if jsonErr != nil {
		log.Printf("Error saat marshal JSON response: %v", jsonErr)
		http.Error(w, "Internal Server Error saat membuat respons JSON", http.StatusInternalServerError)
		return
	}
	_, writeErr := w.Write(jsonResponse)
	if writeErr != nil {
		log.Printf("Error saat menulis JSON response: %v", writeErr)
	}
}

// Fungsi toTitleCase (tetap sama)
func toTitleCase(input string) string {
	words := strings.Fields(input)
	result := make([]string, len(words))
	for i, word := range words {
		if len(word) == 0 {
			continue
		}
		firstChar := strings.ToUpper(string(word[0]))
		restOfWord := ""
		if len(word) > 1 {
			restOfWord = strings.ToLower(word[1:])
		}
		result[i] = firstChar + restOfWord
	}
	return strings.Join(result, " ")
}