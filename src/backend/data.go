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

// ElementImage struct tidak lagi digunakan untuk memuat dari JSON,
// tetapi masih bisa berguna jika Anda ingin memvalidasi keberadaan file gambar secara internal.
// type ElementImage struct {
// 	Name     string `json:"name"`
// 	ImageURL string `json:"imageURL"` // Tidak lagi digunakan untuk URL eksternal
// }

var (
	recipeMap       map[string][]Recipe
	allElementNames map[string]bool
	// imageMap tidak lagi menyimpan URL eksternal, jadi bisa dihapus atau diubah fungsinya.
	// Untuk saat ini, kita akan menghapusnya dari pemuatan data dan imageHandler akan membuat path langsung.
	// imageMap map[string]string

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

		// Tidak perlu lagi memuat element_images_urls.json
		// tempImages, err := loadImages(filepath.Join(dataDir, "element_images_urls.json"))
		// if err != nil {
		// 	loadDataErr = fmt.Errorf("gagal memuat gambar: %w", err)
		// 	return
		// }
		// fmt.Printf("Berhasil memuat %d data URL gambar.\n", len(tempImages))

		fmt.Println("Memproses data resep ke dalam struktur map...")
		processRecipesToMaps(tempRecipes) // Hanya proses resep
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

// Fungsi loadImages tidak lagi diperlukan
// func loadImages(filePath string) ([]ElementImage, error) { ... }

func processRecipesToMaps(recipes []Recipe) {
	recipeMap = make(map[string][]Recipe)
	allElementNames = make(map[string]bool)
	// imageMap tidak diisi dari JSON lagi
	// imageMap = make(map[string]string)

	for _, r := range recipes {
		recipeMap[r.Result] = append(recipeMap[r.Result], r)
		allElementNames[r.Result] = true
		allElementNames[r.Ingredient1] = true
		allElementNames[r.Ingredient2] = true
	}

	// Tidak ada lagi pemrosesan gambar di sini

	baseElements := []string{"Air", "Earth", "Fire", "Water"} // Pastikan elemen dasar tetap ada
	for _, base := range baseElements {
		allElementNames[base] = true
	}

	fmt.Printf("Total elemen unik yang teridentifikasi dari resep: %d\n", len(allElementNames))
}

func GetRecipeMap() map[string][]Recipe {
	return recipeMap
}

// GetImageMap tidak lagi relevan untuk URL eksternal.
// Jika Anda butuh cara untuk memeriksa keberadaan file gambar lokal,
// itu bisa diimplementasikan terpisah atau di dalam imageHandler.
// func GetImageMap() map[string]string {
// 	return imageMap
// }

func GetAllElementNames() map[string]bool {
	return allElementNames
}

func IsElementExists(name string) bool {
	_, exists := allElementNames[name]
	return exists
}