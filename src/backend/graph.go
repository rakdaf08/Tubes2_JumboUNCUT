// src/backend/graph.go
package main

import (
	"fmt"
	"sync"
)

var (
	alchemyGraph map[string][]Recipe

	buildGraphOnce sync.Once
)

func BuildGraph(inputRecipeMap map[string][]Recipe) {
	buildGraphOnce.Do(func() { // Hanya jalankan sekali
		fmt.Println("Membangun struktur graf dari data resep...")
		alchemyGraph = make(map[string][]Recipe)

		for _, recipes := range inputRecipeMap {
			for _, recipe := range recipes {
				alchemyGraph[recipe.Ingredient1] = append(alchemyGraph[recipe.Ingredient1], recipe)
				alchemyGraph[recipe.Ingredient2] = append(alchemyGraph[recipe.Ingredient2], recipe)
			}
		}
		fmt.Printf("Graf selesai dibangun. Jumlah node (elemen bahan) dalam graf: %d\n", len(alchemyGraph))
	})
}

func GetAlchemyGraph() map[string][]Recipe {
	return alchemyGraph
}
