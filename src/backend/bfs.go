// src/backend/bfs.go
package main

import (
	"container/list"
	"errors"
	"fmt"
)

// FindPathBFS (fungsi ini tetap sama)
func FindPathBFS(targetElement string) ([]Recipe, int, error) {
	fmt.Printf("Mencari jalur BFS ke: %s\n", targetElement)
	graph := GetAlchemyGraph()
	if graph == nil {
		return nil, 0, errors.New("graf alkimia belum diinisialisasi")
	}

	queue := list.New()
	visited := make(map[string]bool)
	parent := make(map[string]Recipe) // Tipe map[string]Recipe
	nodesVisitedCount := 0

	baseElements := []string{"Air", "Earth", "Fire", "Water"}

	for _, base := range baseElements {
		if base == targetElement {
			return []Recipe{}, 0, nil
		}
		if _, exists := graph[base]; exists || isBaseElement(base) {
			if !visited[base] {
				queue.PushBack(base)
				visited[base] = true
			}
		}
	}

	for queue.Len() > 0 {
		queueElement := queue.Front()
		currentElement := queueElement.Value.(string)
		queue.Remove(queueElement)
		nodesVisitedCount++

		recipesInvolvingCurrent := graph[currentElement]

		for _, recipe := range recipesInvolvingCurrent {
			var otherIngredient string
			if recipe.Ingredient1 == currentElement {
				otherIngredient = recipe.Ingredient2
			} else {
				otherIngredient = recipe.Ingredient1
			}

			if visited[otherIngredient] {
				result := recipe.Result
				if !visited[result] {
					visited[result] = true
					parent[result] = recipe // Simpan resep pertama
					queue.PushBack(result)

					if result == targetElement {
						fmt.Printf("BFS: Target '%s' ditemukan!\n", targetElement)
						path := reconstructPathRevised(parent, targetElement) // Gunakan map[string]Recipe
						return path, nodesVisitedCount, nil
					}
				}
			}
		}
	}

	fmt.Printf("BFS: Target '%s' tidak dapat ditemukan.\n", targetElement)
	return nil, nodesVisitedCount, fmt.Errorf("jalur ke elemen '%s' tidak ditemukan", targetElement)
}


// FindMultiplePathsBFS (dengan perbaikan tipe parent)
func FindMultiplePathsBFS(targetElement string, maxRecipes int) ([][]Recipe, int, error) {
	fmt.Printf("Mencari %d jalur BFS ke: %s\n", maxRecipes, targetElement)

	graph := GetAlchemyGraph()
	if graph == nil {
		return nil, 0, errors.New("graf alkimia belum diinisialisasi")
	}
    if maxRecipes <= 0 {
        return nil, 0, errors.New("jumlah resep minimal harus 1")
    }


	queue := list.New()
	visited := make(map[string]bool)
	// --- PERUBAHAN TIPE PARENT DI SINI ---
	parent := make(map[string]Recipe) // Kembali ke map[string]Recipe
	// ------------------------------------
	nodesVisitedCount := 0

	allFoundPaths := [][]Recipe{}
	foundCount := 0
    foundAtLevel := -1 // Level pertama kali target ditemukan

	baseElements := []string{"Air", "Earth", "Fire", "Water"}

	for _, base := range baseElements {
		if base == targetElement {
			return [][]Recipe{}, 0, nil
		}
		if _, exists := graph[base]; exists || isBaseElement(base) {
			if !visited[base] {
				queue.PushBack(base)
				visited[base] = true
			}
		}
	}

	level := 0
	elementsInCurrentLevel := queue.Len()
	elementsInNextLevel := 0

	for queue.Len() > 0 {
		// Jika sudah menemukan cukup path DAN sudah melewati level penemuan pertama, berhenti
        if foundCount >= maxRecipes && level > foundAtLevel && foundAtLevel != -1 {
             fmt.Println("BFS Multiple: Batas resep tercapai dan level terpendek sudah selesai.")
             break
        }

		queueElement := queue.Front()
		currentElement := queueElement.Value.(string)
		queue.Remove(queueElement)
		nodesVisitedCount++
		elementsInCurrentLevel--

		recipesInvolvingCurrent := graph[currentElement]

		for _, recipe := range recipesInvolvingCurrent {
			var otherIngredient string
			if recipe.Ingredient1 == currentElement {
				otherIngredient = recipe.Ingredient2
			} else {
				otherIngredient = recipe.Ingredient1
			}

			// Cek apakah bahan lainnya sudah pernah dikunjungi (di level ini atau sebelumnya)
			if visited[otherIngredient] {
				result := recipe.Result

                // Jika hasil belum pernah dikunjungi SAMA SEKALI, baru enqueue dan set parent
				if !visited[result] {
					visited[result] = true
					// --- PERUBAHAN PENGISIAN PARENT ---
					parent[result] = recipe // Simpan resep PERTAMA yang menemukannya
                    // ---------------------------------
					queue.PushBack(result)
					elementsInNextLevel++
				}

				// Cek apakah hasil adalah target, meskipun sudah pernah divisit
				// (karena bisa jadi ada jalur lain di level yang sama)
				if result == targetElement {
                    // Hanya proses jika kita masih mencari ATAU ini adalah level pertama penemuan
                    if foundCount < maxRecipes && (foundAtLevel == -1 || level == foundAtLevel) {
                        if foundAtLevel == -1 {
                             foundAtLevel = level // Catat level pertama kali ditemukan
                        }

                        // Rekonstruksi path menggunakan resep INI sebagai langkah terakhir
                        // Kita perlu cara untuk memasukkan resep ini ke map parent *sementara*
                        // agar reconstructPathRevised bekerja untuk jalur ini.
                        // Atau, modifikasi reconstruct agar menerima resep terakhir.

                        // Pendekatan Sementara: Buat salinan map parent dan tambahkan/update resep terakhir
                        tempParent := make(map[string]Recipe)
                        for k, v := range parent {
                            tempParent[k] = v
                        }
                        tempParent[targetElement] = recipe // Pastikan resep terakhir ini yang dipakai

                        currentPath := reconstructPathRevised(tempParent, targetElement) // Gunakan tempParent

                        // TODO: Cek apakah path ini unik jika diperlukan
                        if len(currentPath) > 0 {
                            allFoundPaths = append(allFoundPaths, currentPath)
                            foundCount++
                            fmt.Printf("BFS Multiple: Path ke-%d ditemukan (target: %s, level: %d)\n", foundCount, targetElement, level)
                            if foundCount >= maxRecipes {
                                // Jangan langsung break loop resep, biarkan cek resep lain di level yg sama
                            }
                        }
                    }
				}
			}
		} // End loop resep

		// Pindah level
		if elementsInCurrentLevel == 0 {
			level++
			elementsInCurrentLevel = elementsInNextLevel
			elementsInNextLevel = 0
            // Jika sudah melewati level penemuan pertama, dan sudah cukup, break loop utama
            if foundAtLevel != -1 && level > foundAtLevel && foundCount >= maxRecipes {
                 fmt.Println("BFS Multiple: Selesai memproses level terpendek.")
                 break
            }
		}

	} // End loop queue

	if foundCount == 0 {
		fmt.Printf("BFS Multiple: Target '%s' tidak dapat ditemukan.\n", targetElement)
		return nil, nodesVisitedCount, fmt.Errorf("jalur ke elemen '%s' tidak ditemukan", targetElement)
	}

	return allFoundPaths, nodesVisitedCount, nil
}


// reconstructPathRevised (fungsi ini tetap sama, mengharapkan map[string]Recipe)
func reconstructPathRevised(parent map[string]Recipe, target string) []Recipe {
    path := list.New()
    queue := list.New()
    queue.PushBack(target)
    processedForPath := make(map[string]bool)
	processedForPath[target]=true

    for queue.Len() > 0 {
        queueEl := queue.Front()
        current := queueEl.Value.(string)
        queue.Remove(queueEl)

        recipe, exists := parent[current]
        if !exists {
            continue
        }

        path.PushFront(recipe)

        if _, processed := processedForPath[recipe.Ingredient1]; !processed {
             if _, parentExists := parent[recipe.Ingredient1]; parentExists || isBaseElement(recipe.Ingredient1) {
                queue.PushBack(recipe.Ingredient1)
                processedForPath[recipe.Ingredient1] = true
             }
        }
		if _, processed := processedForPath[recipe.Ingredient2]; !processed {
             if _, parentExists := parent[recipe.Ingredient2]; parentExists || isBaseElement(recipe.Ingredient2) {
                queue.PushBack(recipe.Ingredient2)
                processedForPath[recipe.Ingredient2] = true
             }
        }
    }

    finalPath := make([]Recipe, 0, path.Len())
    for e := path.Front(); e != nil; e = e.Next() {
        finalPath = append(finalPath, e.Value.(Recipe))
    }
    return finalPath
}

// isBaseElement (fungsi ini tetap sama)
func isBaseElement(name string) bool {
	baseElements := []string{"Air", "Earth", "Fire", "Water"}
	for _, base := range baseElements {
		if name == base {
			return true
		}
	}
	return false
}
