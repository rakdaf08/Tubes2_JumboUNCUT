// src/backend/data.go
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type Recipe struct {
	Result      string `json:"result"`
	Ingredient1 string `json:"ingredient1"`
	Ingredient2 string `json:"ingredient2"`
}

var (
	recipeMap       map[string][]Recipe
	allElementNames map[string]bool

	bfsPathCache = make(map[string][]Recipe)
	loadDataOnce sync.Once
	loadDataErr  error
)

func InitData(dataDir string) error {
	loadDataOnce.Do(func() {
		fmt.Println("Memulai pemuatan data awal dari direktori:", dataDir)

		tempRecipes, err := loadRecipes(filepath.Join(dataDir, "recipes_final_filtered.json"))
		if err != nil {
			loadDataErr = fmt.Errorf("gagal memuat resep: %w", err)
			return
		}
		fmt.Printf("Berhasil memuat %d data resep.\n", len(tempRecipes))

		fmt.Println("Memproses data resep ke dalam struktur map...")
		processRecipesToMaps(tempRecipes)
		fmt.Println("Selesai memproses data resep.")
	})
	return loadDataErr
}

func loadRecipes(filePath string) ([]Recipe, error) {
	fmt.Printf("Membaca file resep: %s\n", filePath)
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca file %s: %w", filePath, err)
	}
	var recipes []Recipe
	err = json.Unmarshal(bytes, &recipes)
	if err != nil {
		return nil, fmt.Errorf("gagal unmarshal JSON resep dari %s: %w", filePath, err)
	}
	return recipes, nil
}

func processRecipesToMaps(recipes []Recipe) {
	recipeMap = make(map[string][]Recipe)
	allElementNames = make(map[string]bool)

	for _, r := range recipes {
		recipeMap[r.Result] = append(recipeMap[r.Result], r)
		allElementNames[r.Result] = true
		allElementNames[r.Ingredient1] = true
		allElementNames[r.Ingredient2] = true
	}

	baseElements := []string{"Air", "Earth", "Fire", "Water"}
	for _, base := range baseElements {
		allElementNames[base] = true
	}

	fmt.Printf("Total elemen unik yang teridentifikasi dari resep: %d\n", len(allElementNames))
}

func GetRecipeMap() map[string][]Recipe {
	return recipeMap
}

func GetAllElementNames() map[string]bool {
	return allElementNames
}

func IsElementExists(name string) bool {
	_, exists := allElementNames[name]
	return exists
}
