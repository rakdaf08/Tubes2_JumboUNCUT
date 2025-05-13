package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type ElementImage struct {
	Name     string `json:"name"`
	ImageURL string `json:"imageURL"`
}

const targetURL = "https://little-alchemy.fandom.com/wiki/Elements_(Little_Alchemy_2)#Tier_15_elements"

func getValidImageURL(imgTag *goquery.Selection) (string, bool) {
	imgSrc, exists := imgTag.Attr("data-src")
	if exists && !strings.HasPrefix(imgSrc, "data:") {
		return imgSrc, true
	}
	imgSrc, exists = imgTag.Attr("src")
	if exists && !strings.HasPrefix(imgSrc, "data:") {
		return imgSrc, true
	}
	return "", false
}

func RunScraping() {
	if targetURL == "URL_WEBSITE_TARGET_ANDA_DI_SINI" {
		log.Fatal("Error: Anda belum mengganti placeholder targetURL di dalam kode!")
	}

	dataDir := "data"
	if err := os.MkdirAll(dataDir, os.ModePerm); err != nil {
		log.Fatalf("Error membuat direktori '%s': %v", dataDir, err)
	}
	fmt.Printf("Memastikan direktori '%s' ada.\n", dataDir)

	fmt.Println("Memulai proses scraping dari:", targetURL)

	res, err := http.Get(targetURL)
	if err != nil {
		log.Fatalf("Error GET request: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("Error status code: %d", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatalf("Error membaca HTML: %v", err)
	}
	fmt.Println("Berhasil memuat dokumen HTML.")

	var allRecipes []Recipe
	var elementImages []ElementImage
	processedElements := make(map[string]bool)

	tableSelector := "table.list-table.col-list.icon-hover"
	fmt.Printf("Mencari tabel dengan selector: '%s'\n", tableSelector)

	doc.Find(tableSelector).Each(func(index int, table *goquery.Selection) {
		fmt.Printf("\nMemproses Tabel ke-%d\n", index+1)
		table.Find("tbody tr").Each(func(i int, row *goquery.Selection) {
			if row.Find("th").Length() > 0 {
				return
			}

			resultCell := row.Find("td:nth-child(1)")
			resultNameLink := resultCell.Find("a")
			resultName := strings.TrimSpace(resultNameLink.Text())
			if resultName == "" {
				return
			}

			fmt.Printf("  Memproses Elemen: %s\n", resultName)

			if _, processed := processedElements[resultName]; !processed {
				imgResultSelector := "span > span > a > img"
				imgURL := ""
				resultCell.Find(imgResultSelector).First().Each(func(_ int, imgTag *goquery.Selection) {
					validURL, isValid := getValidImageURL(imgTag)
					if isValid {
						imgURL = validURL
					}
				})
				if imgURL != "" {
					elementImages = append(elementImages, ElementImage{Name: resultName, ImageURL: imgURL})
					processedElements[resultName] = true
					fmt.Printf("    -> URL Gambar Hasil ditemukan: %s\n", imgURL)
				} else {
					fmt.Printf("    -> Peringatan: Tidak ditemukan URL gambar valid untuk hasil '%s'.\n", resultName)
					processedElements[resultName] = true
				}
			}

			recipesCell := row.Find("td:nth-child(2)")
			if recipesCell.Find("ul").Length() == 0 || strings.Contains(strings.ToLower(recipesCell.Text()), "available from the start") || strings.Contains(strings.ToLower(recipesCell.Text()), "does not have any recipes") {
				return
			}

			liSelector := "ul > li"
			recipesCell.Find(liSelector).Each(func(j int, li *goquery.Selection) {
				var ingredientNames []string
				var ingredientImageURLs []string

				nameIngredientSelector := "a"
				li.Find(nameIngredientSelector).Each(func(k int, nameLink *goquery.Selection) {
					ingName := strings.TrimSpace(nameLink.Text())
					if ingName != "" && ingName != "+" && len(ingName) > 1 {
						ingredientNames = append(ingredientNames, ingName)
					}
				})

				imgIngredientSelector := "span > span > a > img"
				li.Find(imgIngredientSelector).Each(func(k int, imgTag *goquery.Selection) {
					imgURL, isValid := getValidImageURL(imgTag)
					if isValid {
						ingredientImageURLs = append(ingredientImageURLs, imgURL)
					}
				})

				if len(ingredientNames) == 2 {
					bahan1 := ingredientNames[0]
					bahan2 := ingredientNames[1]

					fmt.Printf("    -> Resep ke-%d: %s + %s\n", j+1, bahan1, bahan2)
					recipe := Recipe{Result: resultName, Ingredient1: bahan1, Ingredient2: bahan2}
					allRecipes = append(allRecipes, recipe)

					var imgURL1, imgURL2 string
					if len(ingredientImageURLs) >= 1 {
						imgURL1 = ingredientImageURLs[0]
					}
					if len(ingredientImageURLs) == 2 {
						imgURL2 = ingredientImageURLs[1]
					}

					if _, processed := processedElements[bahan1]; !processed && imgURL1 != "" {
						elementImages = append(elementImages, ElementImage{Name: bahan1, ImageURL: imgURL1})
						processedElements[bahan1] = true
						fmt.Printf("      -> URL Gambar Bahan 1 ditemukan: %s\n", imgURL1)
					}
					if _, processed := processedElements[bahan2]; !processed && imgURL2 != "" {
						elementImages = append(elementImages, ElementImage{Name: bahan2, ImageURL: imgURL2})
						processedElements[bahan2] = true
						fmt.Printf("      -> URL Gambar Bahan 2 ditemukan: %s\n", imgURL2)
					}

				} else {
					fmt.Printf("    -> Peringatan: Gagal memproses resep ke-%d untuk %s. Jumlah Nama Bahan: %d (%v)\n",
						j+1, resultName, len(ingredientNames), ingredientNames)
				}
			})
		})
	})

	fmt.Printf("\nTotal resep tekstual yang berhasil di-scrape: %d\n", len(allRecipes))
	fmt.Printf("Total pemetaan gambar elemen unik yang ditemukan: %d\n", len(elementImages))

	if len(allRecipes) > 0 {
		recipeData, err := json.MarshalIndent(allRecipes, "", "  ")
		if err != nil {
			log.Fatalf("Error marshal JSON Resep: %v", err)
		}
		recipeFileName := filepath.Join(dataDir, "recipes_scraped.json")
		err = os.WriteFile(recipeFileName, recipeData, 0644)
		if err != nil {
			log.Fatalf("Error menulis JSON Resep ke file '%s': %v", recipeFileName, err)
		}
		fmt.Printf("Sukses! Data resep tekstual telah disimpan ke %s\n", recipeFileName)
	} else {
		fmt.Println("Tidak ada resep tekstual yang di-scrape untuk disimpan.")
	}

	if len(elementImages) > 0 {
		imageData, err := json.MarshalIndent(elementImages, "", "  ")
		if err != nil {
			log.Fatalf("Error marshal JSON Gambar: %v", err)
		}
		imageFileName := filepath.Join(dataDir, "element_images_urls.json")
		err = os.WriteFile(imageFileName, imageData, 0644)
		if err != nil {
			log.Fatalf("Error menulis JSON Gambar ke file '%s': %v", imageFileName, err)
		}
		fmt.Printf("Sukses! Data URL gambar elemen telah disimpan ke %s\n", imageFileName)
	} else {
		fmt.Println("Tidak ada data URL gambar elemen yang di-scrape untuk disimpan.")
	}
}
