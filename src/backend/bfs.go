// src/backend/bfs.go
package main

import (
	"container/list"
	"errors"
	"fmt"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
)

var debugMode = false // Set to true for debug output

func debugPrintf(format string, args ...interface{}) {
	if debugMode {
		fmt.Printf(format, args...)
	}
}

// Using a constant slice for better performance
var baseElements = []string{"Air", "Earth", "Fire", "Water"}

// Precomputed map for faster lookup
var baseElementMap = map[string]bool{
	"Air":   true,
	"Earth": true,
	"Fire":  true,
	"Water": true,
}

func isBaseElement(name string) bool {
	return baseElementMap[name]
}

// Cache for paths to avoid recalculating
var (
	bfsPathCacheMutex sync.RWMutex
)

// FindPathBFS finds the single shortest path to targetElement using BFS
// This function is completely deterministic and does not use concurrency
func FindPathBFS(targetElement string) ([]Recipe, int, error) {
	debugPrintf("Finding BFS shortest path to: %s\n", targetElement)
	graph := GetAlchemyGraph()
	if graph == nil {
		return nil, 0, errors.New("alchemy graph not initialized")
	}

	// Check cache with read lock first
	bfsPathCacheMutex.RLock()
	if path, exists := bfsPathCache[targetElement]; exists {
		bfsPathCacheMutex.RUnlock()
		debugPrintf("BFS Cache: Path to '%s' found in cache.\n", targetElement)
		return path, 0, nil
	}
	bfsPathCacheMutex.RUnlock()

	// Base element check early
	if isBaseElement(targetElement) {
		return []Recipe{}, 0, nil
	}

	queue := list.New()
	visited := make(map[string]bool, 1000)        // Pre-allocate space for better performance
	elementVisited := make(map[string]bool, 1000) // Track which elements we've visited
	recipeParent := make(map[string]Recipe)
	discovered := make(map[string]bool, 1000) // Elements we know how to make
	nodesVisitedCount := 0

	// Depth tracking for more efficient path building
	depth := make(map[string]int)

	// To ensure deterministic behavior, always process elements in the same order
	// Initialize with base elements in alphabetical order
	sortedBaseElements := make([]string, len(baseElements))
	copy(sortedBaseElements, baseElements)
	sort.Strings(sortedBaseElements)

	for _, base := range sortedBaseElements {
		elementVisited[base] = true
		discovered[base] = true // Base elements are already available
		queue.PushBack(base)
		depth[base] = 0
		debugPrintf("Enqueue base element: %s\n", base)
	}

	// BFS traversal
	for queue.Len() > 0 {
		currentElement := queue.Remove(queue.Front()).(string)
		currentDepth := depth[currentElement]
		debugPrintf("Dequeue: %s at depth %d\n", currentElement, currentDepth)
		nodesVisitedCount++

		// Get all recipes that use currentElement
		combinableRecipes := graph[currentElement]
		if len(combinableRecipes) == 0 {
			continue // Skip elements with no possible combinations
		}

		// For deterministic behavior, process discovered elements in alphabetical order
		discoveredElementsList := make([]string, 0, len(discovered))
		for element := range discovered {
			discoveredElementsList = append(discoveredElementsList, element)
		}
		sort.Strings(discoveredElementsList)

		// Check all possible combinations with other discovered elements
		for _, otherElement := range discoveredElementsList {
			// Skip checking combinations we've already tried
			pairKey := getPairKey(currentElement, otherElement)
			if visited[pairKey] {
				continue
			}
			visited[pairKey] = true

			// Get all recipes that combine current with other element
			// For deterministic behavior, ensure recipes are sorted consistently
			recipes := getRecipes(currentElement, otherElement)

			for _, recipe := range recipes {
				result := recipe.Result

				// If we haven't discovered this element yet
				if !discovered[result] {
					discovered[result] = true
					recipeParent[result] = recipe
					depth[result] = currentDepth + 1

					// If found target, reconstruct path and cache it
					if result == targetElement {
						debugPrintf("Target '%s' found!\n", targetElement)
						path := buildRecipePath(recipeParent, targetElement, discovered, depth)

						// Cache with write lock
						bfsPathCacheMutex.Lock()
						bfsPathCache[targetElement] = path
						bfsPathCacheMutex.Unlock()

						return path, nodesVisitedCount, nil
					}

					// Add to queue only if not already visited
					if !elementVisited[result] {
						elementVisited[result] = true
						queue.PushBack(result)
						debugPrintf("Enqueue: %s (from %s + %s) at depth %d\n",
							result, currentElement, otherElement, depth[result])
					}
				}
			}
		}
	}

	debugPrintf("Target '%s' cannot be found.\n", targetElement)
	return nil, nodesVisitedCount, fmt.Errorf("path to element '%s' not found", targetElement)
}

// Helper function to create consistent pair keys
func getPairKey(a, b string) string {
	if a > b {
		return b + ":" + a
	}
	return a + ":" + b
}

// Improved and deterministic path building function for the single path finder
func buildRecipePath(recipeParent map[string]Recipe, target string, discovered map[string]bool, depth map[string]int) []Recipe {
	// Create dependency graph for topological sort
	dependencies := make(map[string][]string)
	elementsNeeded := make(map[string]bool)

	// Start with target and work backwards
	queue := list.New()
	queue.PushBack(target)
	elementsNeeded[target] = true

	// Build dependency graph
	for queue.Len() > 0 {
		current := queue.Remove(queue.Front()).(string)

		// Skip base elements
		if isBaseElement(current) {
			continue
		}

		recipe, exists := recipeParent[current]
		if !exists {
			continue // Should never happen with valid data
		}

		// Add dependencies
		ing1, ing2 := recipe.Ingredient1, recipe.Ingredient2
		dependencies[ing1] = append(dependencies[ing1], current)
		dependencies[ing2] = append(dependencies[ing2], current)

		// Queue ingredients in a deterministic order
		ingredients := []string{ing1, ing2}
		sort.Strings(ingredients) // Always process in alphabetical order for determinism

		for _, ingredient := range ingredients {
			if !elementsNeeded[ingredient] && !isBaseElement(ingredient) {
				elementsNeeded[ingredient] = true
				queue.PushBack(ingredient)
			}
		}
	}

	// Build result in correct order
	var result []Recipe
	available := make(map[string]bool)

	// Start with base elements (in deterministic order)
	sortedBaseElements := make([]string, len(baseElements))
	copy(sortedBaseElements, baseElements)
	sort.Strings(sortedBaseElements)
	for _, base := range sortedBaseElements {
		available[base] = true
	}

	// Keep adding recipes until we have the target
	remainingElements := len(elementsNeeded)
	for remainingElements > 0 {
		// Create a deterministic list of elements to check
		candidateElements := make([]string, 0, len(elementsNeeded))
		for element := range elementsNeeded {
			if !available[element] {
				// Get recipe for this element
				recipe, exists := recipeParent[element]
				if !exists {
					continue
				}

				// Check if ingredients are available
				if available[recipe.Ingredient1] && available[recipe.Ingredient2] {
					candidateElements = append(candidateElements, element)
				}
			}
		}

		if len(candidateElements) == 0 {
			// If we reach here, there's a logical error in the path construction
			break
		}

		// Sort candidates for deterministic selection
		sort.SliceStable(candidateElements, func(i, j int) bool {
			// Primary: Sort by depth (shorter paths first)
			if depth[candidateElements[i]] != depth[candidateElements[j]] {
				return depth[candidateElements[i]] < depth[candidateElements[j]]
			}

			// Secondary: Sort by dependency count (more dependencies first)
			depCountI := len(dependencies[candidateElements[i]])
			depCountJ := len(dependencies[candidateElements[j]])
			if depCountI != depCountJ {
				return depCountI > depCountJ
			}

			// Tertiary: Alphabetical (for absolute determinism)
			return candidateElements[i] < candidateElements[j]
		})

		// Select the best candidate
		bestElement := candidateElements[0]

		// Add recipe to result
		recipe := recipeParent[bestElement]
		result = append(result, recipe)
		available[bestElement] = true
		delete(elementsNeeded, bestElement)
		remainingElements--

		// If we have our target, we can stop
		if available[target] {
			break
		}
	}

	return result
}

// Deterministic recipe lookup function
func getRecipes(a, b string) []Recipe {
	graph := GetAlchemyGraph()
	var result []Recipe

	// Try to find the most efficient way to look up recipes
	// If a has fewer recipes than b, iterate through a's recipes
	aRecipes := graph[a]
	bRecipes := graph[b]

	// Choose the smaller list to iterate through for efficiency
	sourceRecipes := aRecipes
	// otherElement := b
	if len(bRecipes) < len(aRecipes) {
		sourceRecipes = bRecipes
		// otherElement := a
	}

	// Find matching recipes
	for _, r := range sourceRecipes {
		if (r.Ingredient1 == a && r.Ingredient2 == b) || (r.Ingredient1 == b && r.Ingredient2 == a) {
			result = append(result, r)
		}
	}

	// Always sort recipes by result name for consistent behavior
	// This is critical for deterministic path finding
	if len(result) > 1 {
		sort.SliceStable(result, func(i, j int) bool {
			return result[i].Result < result[j].Result
		})
	}
	return result
}

// Optimized path identifier generation
func generatePathIdentifier(path []Recipe) string {
	if len(path) == 0 {
		return ""
	}

	// Pre-allocate for better performance
	parts := make([]string, len(path))

	for i, r := range path {
		ing1, ing2 := r.Ingredient1, r.Ingredient2
		if ing1 > ing2 {
			ing1, ing2 = ing2, ing1
		}
		parts[i] = fmt.Sprintf("%s+%s=>%s", ing1, ing2, r.Result)
	}

	// Sort for consistent identification
	sort.Strings(parts)
	return strings.Join(parts, "|")
}

// FindMultiplePathsBFS finds up to 'maxRecipes' different paths to targetElement using multi-threaded BFS
func FindMultiplePathsBFS(targetElement string, maxRecipes int) ([][]Recipe, int, error) {
	fmt.Printf("Finding %d different BFS paths to: %s (Multithreaded)\n", maxRecipes, targetElement)

	graph := GetAlchemyGraph()
	if graph == nil {
		return nil, 0, errors.New("alchemy graph not initialized")
	}
	if maxRecipes <= 0 {
		return nil, 0, errors.New("minimum number of recipes must be 1")
	}
	if isBaseElement(targetElement) {
		return [][]Recipe{}, 0, nil
	}

	// Structures for results and synchronization
	var allFoundPaths [][]Recipe                  // Final result
	addedPathIdentifiers := make(map[string]bool) // For checking path duplication
	var mu sync.Mutex                             // Mutex for thread-safe access
	nodesVisitedCount := atomic.Int32{}           // Atomic counter for visited nodes

	// Worker pool synchronization
	var wg sync.WaitGroup                       // WaitGroup to wait for goroutines
	pathChan := make(chan []Recipe, maxRecipes) // Channel for found paths
	done := atomic.Bool{}                       // Atomic flag for signaling completion

	// Create a shared cache for worker coordination
	sharedVisited := sync.Map{}

	// Helper function to check if we should stop
	shouldStop := func() bool {
		mu.Lock()
		isDone := len(allFoundPaths) >= maxRecipes
		mu.Unlock()
		return isDone || done.Load()
	}

	// Create worker pool for BFS search
	numWorkers := runtime.NumCPU() // Use CPU count for workers
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			// BFS structures per worker
			queue := list.New()
			localVisited := make(map[string]bool)
			parent := make(map[string]Recipe)
			discovered := make(map[string]bool) // Elements we know how to make
			depthMap := make(map[string]int)    // Track element depths

			// Initialize with different starting points for diversity
			startOffset := workerID % len(baseElements)
			for i := 0; i < len(baseElements); i++ {
				idx := (startOffset + i) % len(baseElements)
				base := baseElements[idx]
				queue.PushBack(base)
				localVisited[base] = true
				discovered[base] = true
				depthMap[base] = 0
			}

			// BFS loop
			for queue.Len() > 0 && !shouldStop() {
				// Process next node
				currentElement := queue.Remove(queue.Front()).(string)
				currentDepth := depthMap[currentElement]
				nodesVisitedCount.Add(1)

				// Try all combinations with discovered elements
				for otherElement := range discovered {
					// Generate unique pair key to avoid duplicate work across workers
					pairKey := getPairKey(currentElement, otherElement)

					// Check if this pair has been processed by any worker
					if _, exists := sharedVisited.LoadOrStore(pairKey, true); exists {
						continue
					}

					// Find all recipes combining current with other element
					for _, recipe := range getRecipes(currentElement, otherElement) {
						if shouldStop() {
							return
						}

						result := recipe.Result

						// Try storing this as an alternative path
						tempParent := make(map[string]Recipe)
						for k, v := range parent {
							tempParent[k] = v
						}
						tempParent[result] = recipe

						newDiscovered := discovered[result] // Already discovered?

						// If not discovered yet, add to regular search
						if !newDiscovered {
							parent[result] = recipe
							discovered[result] = true
							depthMap[result] = currentDepth + 1

							if !localVisited[result] {
								localVisited[result] = true
								queue.PushBack(result)
							}
						}

						// If this is our target, try to build a path
						if result == targetElement {
							// Try with the current parent map or our temporary one
							pathMap := parent
							if newDiscovered {
								pathMap = tempParent
							}

							// Attempt to build a valid path
							currentPath := buildValidPath(pathMap, targetElement)
							if len(currentPath) > 0 {
								pathID := generatePathIdentifier(currentPath)

								mu.Lock()
								if !addedPathIdentifiers[pathID] && len(allFoundPaths) < maxRecipes {
									// Create a copy to avoid race conditions
									pathToAdd := make([]Recipe, len(currentPath))
									copy(pathToAdd, currentPath)

									// Save this path
									allFoundPaths = append(allFoundPaths, pathToAdd)
									addedPathIdentifiers[pathID] = true

									fmt.Printf("Worker %d: Unique path #%d found (target: %s)\n",
										workerID, len(allFoundPaths), targetElement)

									// Send path to channel
									select {
									case pathChan <- pathToAdd:
									default: // Non-blocking
									}

									// Set done flag if we have enough paths
									if len(allFoundPaths) >= maxRecipes {
										done.Store(true)
									}
								}
								mu.Unlock()
							}
						}
					}
				}
			}
		}(w)
	}

	// Goroutine to collect results
	go func() {
		wg.Wait()
		close(pathChan)
	}()

	// Wait for all paths to be found or all workers to finish
	for range pathChan {
		// We're already saving paths to allFoundPaths, this is just for synchronization
	}

	// Prepare final result
	mu.Lock()
	result := make([][]Recipe, len(allFoundPaths))
	copy(result, allFoundPaths)
	mu.Unlock()

	if len(result) == 0 && !isBaseElement(targetElement) {
		fmt.Printf("BFS Multiple: Target '%s' cannot be found.\n", targetElement)
		return nil, int(nodesVisitedCount.Load()), fmt.Errorf("path to element '%s' not found", targetElement)
	}

	return result, int(nodesVisitedCount.Load()), nil
}

// buildValidPath ensures the created path is valid (all required ingredients are created)
func buildValidPath(parent map[string]Recipe, target string) []Recipe {
	// Collect all needed recipes
	recipesNeeded := make(map[string]Recipe)

	// Start from target
	queue := list.New()
	queue.PushBack(target)
	processed := make(map[string]bool)
	processed[target] = true

	// Process all elements
	for queue.Len() > 0 {
		current := queue.Remove(queue.Front()).(string)

		// Skip base elements
		if isBaseElement(current) {
			continue
		}

		// Get recipe for this element
		recipe, exists := parent[current]
		if !exists {
			return []Recipe{} // Invalid path if any element has no recipe
		}

		// Add recipe to list
		recipesNeeded[current] = recipe

		// Queue ingredients for processing
		for _, ingredient := range []string{recipe.Ingredient1, recipe.Ingredient2} {
			if !processed[ingredient] && !isBaseElement(ingredient) {
				queue.PushBack(ingredient)
				processed[ingredient] = true
			}
		}
	}

	// Build path in correct order
	var path []Recipe
	availableElements := make(map[string]bool)

	// Start with base elements
	for _, base := range baseElements {
		availableElements[base] = true
	}

	// Add recipes until we have the target
	for len(recipesNeeded) > 0 && !availableElements[target] {
		// Find recipe where we have all ingredients
		added := false
		for element, recipe := range recipesNeeded {
			if availableElements[recipe.Ingredient1] && availableElements[recipe.Ingredient2] {
				// Add recipe to path
				path = append(path, recipe)
				availableElements[recipe.Result] = true

				// Remove from needed recipes
				delete(recipesNeeded, element)
				added = true
				break
			}
		}

		// If no recipe could be added, path is invalid
		if !added {
			return []Recipe{} // Invalid path
		}
	}

	return path
}
