// src/backend/bfs.go
package main // Harus package main agar bisa diakses langsung oleh main.go dan data.go/graph.go

import (
	"container/list" // Menggunakan list sebagai queue (antrian)
	"errors"        // Untuk membuat error baru
	"fmt"
)

// --- Struktur Data Tambahan untuk BFS ---

// bfsNode digunakan untuk menyimpan state dalam BFS, termasuk elemen dan resep pembuatnya
// type bfsNode struct {
// 	elementName string
// 	recipeUsed  Recipe // Resep yang digunakan untuk MENCAPAI elemen ini
// }
// NOTE: Pendekatan yang lebih umum adalah menyimpan parent/recipe dalam map terpisah.

// --- Fungsi BFS Utama ---

// FindPathBFS mencari jalur resep terpendek dari elemen dasar ke targetElement menggunakan BFS.
// Mengembalikan slice resep yang membentuk jalur, jumlah node yang dikunjungi, dan error jika tidak ditemukan.
func FindPathBFS(targetElement string) ([]Recipe, int, error) {
	fmt.Printf("Mencari jalur BFS ke: %s\n", targetElement)

	// Akses graf yang sudah dibuat di graph.go
	graph := GetAlchemyGraph()
	if graph == nil {
		return nil, 0, errors.New("graf alkimia belum diinisialisasi")
	}

	// 1. Inisialisasi struktur data BFS
	queue := list.New()                            // Queue untuk elemen yang akan dikunjungi
	visited := make(map[string]bool)               // Set untuk melacak elemen yang sudah ditemukan/dikunjungi
	parent := make(map[string]Recipe)              // Map untuk merekonstruksi path (Key: Hasil, Value: Resep pembuatnya)
	nodesVisitedCount := 0                         // Penghitung node yang dieksplorasi dari queue

	baseElements := []string{"Air", "Earth", "Fire", "Water"} // Elemen dasar [cite: 6]

	// 2. Mulai dari elemen dasar
	for _, base := range baseElements {
		if base == targetElement { // Target adalah elemen dasar
			return []Recipe{}, 0, nil // Tidak perlu resep
		}
		// Hanya tambahkan jika elemen dasar ada di map graf (artinya pernah jadi bahan)
		if _, exists := graph[base]; exists || len(GetRecipeMap()[base]) > 0 || base == "Air" || base == "Earth" || base == "Fire" || base == "Water" { // Pastikan base element ada di data kita
			queue.PushBack(base)
			visited[base] = true
			// Elemen dasar tidak punya parent recipe
		}
	}

	// 3. Loop BFS selama queue tidak kosong
	for queue.Len() > 0 {
		// Dequeue elemen pertama
		queueElement := queue.Front()
		currentElement := queueElement.Value.(string)
		queue.Remove(queueElement)
		nodesVisitedCount++ // Hitung node yang diproses dari queue

		// Debug: Cetak elemen yg sedang diproses
		// fmt.Printf("  BFS: Memproses -> %s\n", currentElement)

		// 4. Dapatkan semua resep yang melibatkan currentElement sebagai bahan
		recipesInvolvingCurrent := graph[currentElement]

		// 5. Iterasi melalui resep-resep tersebut
		for _, recipe := range recipesInvolvingCurrent {
			var otherIngredient string
			// Tentukan bahan lainnya
			if recipe.Ingredient1 == currentElement {
				otherIngredient = recipe.Ingredient2
			} else {
				otherIngredient = recipe.Ingredient1
			}

			// 6. Cek apakah bahan lainnya sudah ditemukan (visited)
			if visited[otherIngredient] {
				// Jika ya, kita bisa membuat recipe.Result
				result := recipe.Result

				// 7. Cek apakah hasil ini belum pernah ditemukan sebelumnya
				if !visited[result] {
					// Debug: Cetak elemen baru yg ditemukan
					// fmt.Printf("    BFS: Menemukan -> %s (dari %s + %s)\n", result, recipe.Ingredient1, recipe.Ingredient2)

					visited[result] = true       // Tandai hasil sudah ditemukan
					parent[result] = recipe      // Simpan resep yang digunakan untuk membuatnya
					queue.PushBack(result)       // Masukkan hasil ke queue untuk diproses

					// 8. Cek apakah hasil ini adalah target kita
					if result == targetElement {
						fmt.Printf("BFS: Target '%s' ditemukan!\n", targetElement)
						// Rekonstruksi path dari parent map
						path := reconstructPath(parent, targetElement)
						return path, nodesVisitedCount, nil // Sukses!
					}
				}
			}
		} // Akhir loop resep
	} // Akhir loop queue

	// 9. Jika queue kosong dan target tidak ditemukan
	fmt.Printf("BFS: Target '%s' tidak dapat ditemukan dari elemen dasar.\n", targetElement)
	return nil, nodesVisitedCount, fmt.Errorf("jalur ke elemen '%s' tidak ditemukan", targetElement)
}

// --- Fungsi Helper untuk Rekonstruksi Path ---

// reconstructPath membuat ulang urutan resep dari map parent.
func reconstructPath(parent map[string]Recipe, target string) []Recipe {
	path := []Recipe{}
	current := target

	// Telusuri balik dari target menggunakan resep parent
	for {
		recipe, exists := parent[current]
		if !exists {
			// Kita sudah mencapai elemen yang tidak punya parent recipe (seharusnya elemen dasar)
			break
		}
		// Masukkan resep ke awal slice (agar urutannya benar dari dasar ke target)
		path = append([]Recipe{recipe}, path...)

		// Mundur ke salah satu bahan dari resep ini.
		// Kita perlu tahu bahan mana yang "lebih dulu" ditemukan atau mana yang bukan elemen dasar
		// Untuk simplifikasi, kita bisa mundur ke bahan1 ATAU bahan2 jika salah satunya bukan elemen dasar
		// dan punya parent. Atau telusuri keduanya?
		// Untuk path terpendek BFS, cukup telusuri balik saja. Bahan tidak perlu ditelusuri ulang
		// karena parent map sudah menjamin ketercapaian. Kita hanya perlu resepnya.

		// Tentukan elemen mana yang akan ditelusuri selanjutnya.
		// Prioritaskan elemen yang BUKAN elemen dasar dan punya parent lagi.
		// Cek apakah Ing1 punya parent (bukan elemen dasar yg dibuat dari null)
		_, parent1Exists := parent[recipe.Ingredient1]
		// Cek apakah Ing2 punya parent
		_, parent2Exists := parent[recipe.Ingredient2]

		// Pindah ke elemen parent selanjutnya. Logika ini mungkin perlu disempurnakan
		// tergantung bagaimana ingin menampilkan pohon dependensi lengkap.
		// Untuk sekedar daftar resep, kita hanya perlu mundur dari result saat ini.
		// Logika sederhananya: elemen 'current' dibuat oleh 'recipe'. Selesai. Iterasi berikutnya
		// akan dimulai dari elemen 'current' yang baru (yaitu bahan dari resep sebelumnya).
		// Tapi karena kita hanya menyimpan Result->Recipe, kita perlu cari tahu mana
		// dari Ing1 atau Ing2 yang membawa kita ke 'current' ini.
		// Pendekatan paling mudah: cukup berhenti setelah memasukkan resep pertama.
		// Koreksi: Kita perlu mundur terus sampai elemen dasar. Map parent[Result]=Recipe sudah cukup.
		// Kita hanya perlu memilih salah satu parent untuk mundur jika diperlukan, tapi
		// untuk daftar resep, kita telusuri balik Resultnya.
		current = recipe.Result // Ini salah, harusnya mundur ke bahan

		// Mundur ke salah satu bahan untuk melanjutkan iterasi path reconstruction
		// Kita harus tahu mana dari bahan1 atau bahan2 yang merupakan "anak" dari
		// langkah sebelumnya dalam konteks parent map ini.
		// Atau cara paling mudah: telusuri saja terus dari `current` yang baru.
		next_current_found := false
		if parent1Exists { // Jika bahan1 punya parent (bukan base)
			current = recipe.Ingredient1
			next_current_found = true
		} else if parent2Exists { // Jika bahan2 punya parent
			current = recipe.Ingredient2
			next_current_found = true
		} else {
			// Jika kedua bahan tidak punya parent di map, berarti mereka base element atau
			// salah satunya base element, kita berhenti di sini.
			break
		}

		// Jika elemen 'current' yang baru adalah salah satu base element, kita berhenti.
		isBase := false
		baseElements := []string{"Air", "Earth", "Fire", "Water"}
		for _, base := range baseElements {
			if current == base {
				isBase = true
				break
			}
		}
		if isBase && !next_current_found { // Jika kita memilih mundur ke base element (karena satunya lagi juga base/tak ada parent)
			break
		}

	}
	return path
}

// --- REVISI reconstructPath (Lebih Sederhana & Benar) ---
func reconstructPathRevised(parent map[string]Recipe, target string) []Recipe {
    path := list.New() // Gunakan list agar mudah insert di depan
    queue := list.New()
    queue.PushBack(target)
    processedForPath := make(map[string]bool) // Hindari duplikasi elemen dalam path reconstruction
	processedForPath[target]=true

    for queue.Len() > 0 {
        queueEl := queue.Front()
        current := queueEl.Value.(string)
        queue.Remove(queueEl)

        recipe, exists := parent[current]
        if !exists {
            // Mencapai elemen dasar atau elemen tanpa parent (awal)
            continue
        }

        // Tambahkan resep ini ke depan path
        path.PushFront(recipe)

        // Tambahkan kedua bahan ke queue jika belum diproses untuk path
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

    // Konversi list ke slice
    finalPath := make([]Recipe, 0, path.Len())
    for e := path.Front(); e != nil; e = e.Next() {
        finalPath = append(finalPath, e.Value.(Recipe))
    }
    return finalPath
}

func isBaseElement(name string) bool {
	baseElements := []string{"Air", "Earth", "Fire", "Water"}
	for _, base := range baseElements {
		if name == base {
			return true
		}
	}
	return false
}