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
func FindPathBFS(targetElement string) ([]Recipe, int, error) {
	fmt.Printf("Finding BFS shortest path to: %s\n", targetElement)
	graph := GetAlchemyGraph()
	if graph == nil {
		return nil, 0, errors.New("alchemy graph not initialized")
	}

	// Check cache with read lock first
	bfsPathCacheMutex.RLock()
	if path, exists := bfsPathCache[targetElement]; exists {
		bfsPathCacheMutex.RUnlock()
		fmt.Printf("BFS Cache: Path to '%s' found in cache.\n", targetElement)
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

		fmt.Printf("Enqueue base element: %s\n", base)
	}

	// BFS traversal
	for queue.Len() > 0 {
		currentElement := queue.Remove(queue.Front()).(string)
		currentDepth := depth[currentElement]
		fmt.Printf("Dequeue: %s at depth %d\n", currentElement, currentDepth)
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
						fmt.Printf("Target '%s' found!\n", targetElement)
						path := buildRecipePath(recipeParent, targetElement, depth)

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
						fmt.Printf("Enqueue: %s (from %s + %s) at depth %d\n",
							result, currentElement, otherElement, depth[result])
					}
				}
			}
		}
	}
	fmt.Printf("Target '%s' cannot be found.\n", targetElement)
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
func buildRecipePath(recipeParent map[string]Recipe, target string, depth map[string]int) []Recipe {
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

// Modified path identifier generation to focus on unique ingredient combinations
func generatePathIdentifier(path []Recipe) string {
	if len(path) == 0 {
		return ""
	}

	// Create a map of result -> ingredient pair for unique combinations
	resultToIngredients := make(map[string]string)

	for _, r := range path {
		// Always sort ingredients lexicographically to ensure A+B and B+A are treated the same
		ing1, ing2 := r.Ingredient1, r.Ingredient2
		if ing1 > ing2 {
			ing1, ing2 = ing2, ing1
		}
		resultToIngredients[r.Result] = ing1 + "+" + ing2
	}

	// Convert to sorted slice for deterministic output
	parts := make([]string, 0, len(resultToIngredients))
	for result, ingredients := range resultToIngredients {
		parts = append(parts, fmt.Sprintf("%s=>%s", ingredients, result))
	}

	// Sort for consistent identification
	sort.Strings(parts)
	return strings.Join(parts, "|")
}

// Helper function to get a unique recipe key by ingredient combination
func getUniqueRecipeKey(recipe Recipe) string {
	ing1, ing2 := recipe.Ingredient1, recipe.Ingredient2
	if ing1 > ing2 {
		ing1, ing2 = ing2, ing1
	}
	return fmt.Sprintf("%s+%s=>%s", ing1, ing2, recipe.Result)
}

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

	// Count and identify all possible unique recipe combinations for the target
	uniqueRecipeCombos, allCombinations := getAllUniqueRecipeCombinations(targetElement)
	if uniqueRecipeCombos == 0 {
		return nil, 0, fmt.Errorf("element '%s' not found in recipe database", targetElement)
	}

	// Display all available combinations for transparency
	fmt.Printf("Element '%s' can be created from %d unique ingredient combinations:\n",
		targetElement, uniqueRecipeCombos)
	for comboKey, recipe := range allCombinations {
		fmt.Printf("  - %s + %s => %s (key: %s)\n",
			recipe.Ingredient1, recipe.Ingredient2, recipe.Result, comboKey)
	}

	// Adjust maxRecipes to not exceed the actual number of unique combinations
	if uniqueRecipeCombos < maxRecipes {
		fmt.Printf("Adjusting max paths to %d to match available combinations\n", uniqueRecipeCombos)
		maxRecipes = uniqueRecipeCombos
	}

	// If only 1 path is possible, use the simpler BFS algorithm
	if maxRecipes == 1 {
		firstPath, visitCount, err := FindPathBFS(targetElement)
		if err != nil {
			return nil, visitCount, err
		}
		return [][]Recipe{firstPath}, visitCount, nil
	}

	// Track already found unique ingredient combinations for the target
	foundTargetCombinations := make(map[string]bool)
	// Map to track which combinations we're actively searching for
	remainingCombinations := make(map[string]Recipe, uniqueRecipeCombos)
	for comboKey, recipe := range allCombinations {
		remainingCombinations[comboKey] = recipe
	}

	// Structures for results and synchronization
	var allFoundPaths [][]Recipe
	addedPathIdentifiers := make(map[string]bool)
	var mu sync.Mutex
	nodesVisitedCount := atomic.Int32{}

	// Worker pool synchronization
	var wg sync.WaitGroup
	pathChan := make(chan []Recipe, maxRecipes)
	done := atomic.Bool{}

	// Start with the basic path from single-path BFS
	firstPath, _, firstErr := FindPathBFS(targetElement)
	if firstErr == nil && len(firstPath) > 0 {
		pathID := generatePathIdentifier(firstPath)

		// Find the target recipe in the path
		var targetRecipe Recipe
		for _, r := range firstPath {
			if r.Result == targetElement {
				targetRecipe = r
				break
			}
		}

		// Track this specific target combination
		comboKey := getUniqueRecipeKey(targetRecipe)

		mu.Lock()
		allFoundPaths = append(allFoundPaths, firstPath)
		addedPathIdentifiers[pathID] = true
		foundTargetCombinations[comboKey] = true
		delete(remainingCombinations, comboKey)
		mu.Unlock()

		fmt.Printf("Initial path found for %s via FindPathBFS using ingredients: %s + %s\n",
			targetElement, targetRecipe.Ingredient1, targetRecipe.Ingredient2)

		// Send to channel, but don't block
		select {
		case pathChan <- firstPath:
		default:
		}
	}

	// Helper function to check if we should stop
	shouldStop := func() bool {
		mu.Lock()
		isDone := len(allFoundPaths) >= maxRecipes || len(foundTargetCombinations) >= uniqueRecipeCombos
		mu.Unlock()
		return isDone || done.Load()
	}

	// If we still need more paths, launch targeted searches for each remaining combination
	if len(foundTargetCombinations) < uniqueRecipeCombos && len(allFoundPaths) < maxRecipes {
		// Adjust search parameters based on element tier
		numWorkersPerCombo := 3 // Multiple workers per combination for redundancy

		// First, ensure we have at least one worker per remaining combination
		mu.Lock()
		combinationsToSearch := make([]Recipe, 0, len(remainingCombinations))
		for _, recipe := range remainingCombinations {
			combinationsToSearch = append(combinationsToSearch, recipe)
		}
		mu.Unlock()

		// Launch targeted searches for each remaining combination
		for comboIdx, targetRecipe := range combinationsToSearch {
			for w := 0; w < numWorkersPerCombo; w++ {
				if shouldStop() {
					break
				}

				wg.Add(1)
				go func(workerID int, comboIdx int, targetComboRecipe Recipe) {
					defer wg.Done()

					comboKey := getUniqueRecipeKey(targetComboRecipe)

					// Check if this combination has been found while we were setting up
					mu.Lock()
					alreadyFound := foundTargetCombinations[comboKey]
					mu.Unlock()

					if alreadyFound {
						return
					}

					// Create a strategy variant for this worker
					strategyVariant := (workerID + comboIdx) % 5

					fmt.Printf("Worker %d searching for combo %d: %s + %s => %s (strategy: %d)\n",
						workerID, comboIdx,
						targetComboRecipe.Ingredient1, targetComboRecipe.Ingredient2,
						targetComboRecipe.Result, strategyVariant)

					// Run targeted search for this specific combination
					currentPath := findPathForSpecificCombination(
						targetElement,
						targetComboRecipe,
						strategyVariant,
						&nodesVisitedCount,
						shouldStop,
					)

					if len(currentPath) > 0 {
						// Verify this path actually uses the targeted combination
						var foundTargetRecipe Recipe
						for _, r := range currentPath {
							if r.Result == targetElement {
								foundTargetRecipe = r
								break
							}
						}

						// Double-check this is the combination we were looking for
						pathComboKey := getUniqueRecipeKey(foundTargetRecipe)
						if pathComboKey != comboKey {
							fmt.Printf("Warning: Worker %d found wrong combination %s instead of %s\n",
								workerID, pathComboKey, comboKey)
							return
						}

						pathID := generatePathIdentifier(currentPath)

						// Check if this path is new
						mu.Lock()
						isNewCombo := !foundTargetCombinations[pathComboKey]
						isNewPath := !addedPathIdentifiers[pathID]

						if isNewCombo && isNewPath && len(allFoundPaths) < maxRecipes {
							// Create a copy to avoid race conditions
							pathCopy := make([]Recipe, len(currentPath))
							copy(pathCopy, currentPath)

							// Add to results
							allFoundPaths = append(allFoundPaths, pathCopy)
							addedPathIdentifiers[pathID] = true
							foundTargetCombinations[pathComboKey] = true
							delete(remainingCombinations, pathComboKey)

							fmt.Printf("Worker %d: Found path #%d for %s using ingredients: %s + %s (strategy: %d)\n",
								workerID, len(allFoundPaths), targetElement,
								foundTargetRecipe.Ingredient1, foundTargetRecipe.Ingredient2,
								strategyVariant)

							// Send to channel
							select {
							case pathChan <- pathCopy:
							default:
							}

							// Check if we have enough paths or all possible combinations
							if len(allFoundPaths) >= maxRecipes || len(foundTargetCombinations) >= uniqueRecipeCombos {
								done.Store(true)
							}
						}
						mu.Unlock()
					}
				}(w, comboIdx, targetRecipe)
			}
		}

		// If we still need more paths after targeted searches, launch some general workers
		if !shouldStop() {
			additionalWorkers := runtime.NumCPU() * 2

			// Launch additional general workers with diverse strategies
			for w := 0; w < additionalWorkers; w++ {
				wg.Add(1)
				go func(workerID int) {
					defer wg.Done()

					// Create a strategy variant for this worker
					strategyVariant := workerID % 5

					// Worker-local structures
					queue := list.New()
					localVisited := make(map[string]bool)
					parent := make(map[string]Recipe)
					discovered := make(map[string]bool)

					// Initialize with base elements in a worker-specific order
					startOffset := (workerID * 17) % len(baseElements) // Prime number for distribution
					for i := 0; i < len(baseElements); i++ {
						idx := (startOffset + i) % len(baseElements)
						base := baseElements[idx]
						queue.PushBack(base)
						localVisited[base] = true
						discovered[base] = true
					}

					// Track depth of each element for better searching
					depthMap := make(map[string]int)
					for _, base := range baseElements {
						depthMap[base] = 0
					}

					// Keep exploring as long as we need more paths
					for queue.Len() > 0 && !shouldStop() {
						// Get next element to explore
						currentElement := queue.Remove(queue.Front()).(string)
						currentDepth := depthMap[currentElement]
						nodesVisitedCount.Add(1)

						// Periodically check remaining combinations and focus search
						if nodesVisitedCount.Load()%1000 == 0 {
							mu.Lock()
							// If only a few combinations remain, focus on those
							if len(remainingCombinations) > 0 && len(remainingCombinations) <= 3 {
								// Pick a remaining combination to focus on
								var targetCombo Recipe
								for _, recipe := range remainingCombinations {
									targetCombo = recipe
									break
								}
								mu.Unlock()

								// Try to discover the ingredients needed for this combination
								ing1 := targetCombo.Ingredient1
								ing2 := targetCombo.Ingredient2

								// Prioritize discovering these ingredients
								if !discovered[ing1] {
									// Try to find a path to this ingredient
									queue.PushFront(ing1)
								}
								if !discovered[ing2] {
									// Try to find a path to this ingredient
									queue.PushFront(ing2)
								}
							} else {
								mu.Unlock()
							}
						}

						// Get all elements to combine with
						combinableElements := make([]string, 0, len(discovered))
						for elem := range discovered {
							combinableElements = append(combinableElements, elem)
						}

						// Sort differently based on worker strategy
						sortElements(combinableElements, strategyVariant, depthMap, workerID)

						// Try combinations with other elements
						for _, otherElement := range combinableElements {
							// Generate unique pair key
							pairKey := getPairKey(currentElement, otherElement)

							// Skip if already visited by this worker
							if localVisited[pairKey] {
								continue
							}
							localVisited[pairKey] = true

							// Get recipes for this combination
							recipes := getRecipes(currentElement, otherElement)

							// Process each recipe result
							for _, recipe := range recipes {
								if shouldStop() {
									return
								}

								result := recipe.Result
								resultDepth := currentDepth + 1

								// Update the parent and depth tracking
								_, alreadyFound := parent[result]

								// For diversity, sometimes override existing parents
								shouldOverride := false
								if alreadyFound {
									rnd := (int(nodesVisitedCount.Load()) + workerID + int(resultDepth)) % 100
									shouldOverride = rnd < 15 // 15% chance
								}

								// Update parent if new or chosen for override
								if !alreadyFound || shouldOverride {
									parent[result] = recipe
									depthMap[result] = resultDepth
								}

								// Add to discovered elements
								wasNewDiscovery := !discovered[result]
								discovered[result] = true

								// Add to queue if not already visited
								queueIt := wasNewDiscovery
								if wasNewDiscovery || shouldOverride {
									if !localVisited[result] || shouldOverride {
										localVisited[result] = true
										if queueIt {
											queue.PushBack(result)
										}
									}
								}

								// If this is our target, try to build a path
								if result == targetElement {
									// Check if this specific target combination has already been found
									comboKey := getUniqueRecipeKey(recipe)

									mu.Lock()
									alreadyFoundThisCombo := foundTargetCombinations[comboKey]
									mu.Unlock()

									// Skip if we've already found a path using this ingredient combination
									if alreadyFoundThisCombo {
										continue
									}

									// Build a valid path
									currentPath := buildDiversePath(parent, targetElement, workerID)
									if len(currentPath) > 0 {
										// Find the target recipe in the path we just built
										var pathTargetRecipe Recipe
										for _, r := range currentPath {
											if r.Result == targetElement {
												pathTargetRecipe = r
												break
											}
										}

										// Double-check this is actually a unique ingredient combination
										pathComboKey := getUniqueRecipeKey(pathTargetRecipe)

										pathID := generatePathIdentifier(currentPath)

										// Check if this path is new
										mu.Lock()
										isNewCombo := !foundTargetCombinations[pathComboKey]
										isNewPath := !addedPathIdentifiers[pathID]

										if isNewCombo && isNewPath && len(allFoundPaths) < maxRecipes {
											// Create a copy to avoid race conditions
											pathCopy := make([]Recipe, len(currentPath))
											copy(pathCopy, currentPath)

											// Add to results
											allFoundPaths = append(allFoundPaths, pathCopy)
											addedPathIdentifiers[pathID] = true
											foundTargetCombinations[pathComboKey] = true
											delete(remainingCombinations, pathComboKey)

											fmt.Printf("Worker %d: Found path #%d for %s using ingredients: %s + %s (strategy: %d)\n",
												workerID, len(allFoundPaths), targetElement,
												pathTargetRecipe.Ingredient1, pathTargetRecipe.Ingredient2,
												strategyVariant)

											// Send to channel
											select {
											case pathChan <- pathCopy:
											default:
											}

											// Check if we have enough paths or all possible combinations
											if len(allFoundPaths) >= maxRecipes || len(foundTargetCombinations) >= uniqueRecipeCombos {
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
		}
	}

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(pathChan)
	}()

	// Collect paths
	for range pathChan {
		// Paths are already saved to allFoundPaths
	}

	// Prepare final result
	mu.Lock()
	result := make([][]Recipe, len(allFoundPaths))
	copy(result, allFoundPaths)

	// Check if any combinations were never found
	missingCount := len(remainingCombinations)
	if missingCount > 0 {
		fmt.Printf("Warning: %d combinations were never found:\n", missingCount)
		for comboKey, recipe := range remainingCombinations {
			fmt.Printf("  - Missing: %s + %s => %s (key: %s)\n",
				recipe.Ingredient1, recipe.Ingredient2, recipe.Result, comboKey)
		}
	}

	// If we have at least one path but less than requested, that's still a success
	foundCount := len(result)
	foundCombinations := len(foundTargetCombinations)
	mu.Unlock()

	if foundCount == 0 && !isBaseElement(targetElement) {
		fmt.Printf("BFS Multiple: No paths found for '%s'.\n", targetElement)
		return nil, int(nodesVisitedCount.Load()), fmt.Errorf("path to element '%s' not found", targetElement)
	}

	fmt.Printf("BFS Multiple: Found %d unique paths (using %d/%d unique ingredient combinations) for '%s' (requested %d).\n",
		foundCount, foundCombinations, uniqueRecipeCombos, targetElement, maxRecipes)
	return result, int(nodesVisitedCount.Load()), nil
}

func getAllUniqueRecipeCombinations(element string) (int, map[string]Recipe) {
	uniqueCombos := make(map[string]Recipe)

	if isBaseElement(element) {
		return 0, uniqueCombos // Base elements have no recipes
	}

	graph := GetAlchemyGraph()
	if graph == nil {
		return 0, uniqueCombos
	}

	// Check all recipes in the graph
	for _, recipes := range graph {
		for _, recipe := range recipes {
			if recipe.Result == element {
				// Generate unique key for this ingredient combination
				comboKey := getUniqueRecipeKey(recipe)
				uniqueCombos[comboKey] = recipe
			}
		}
	}

	return len(uniqueCombos), uniqueCombos
}

// New function to find a path that ensures a specific combination is used
func findPathForSpecificCombination(targetElement string, targetRecipe Recipe,
	strategyVariant int, nodesVisitedCount *atomic.Int32, shouldStop func() bool) []Recipe {

	// Ensure the ingredients of the target recipe are discovered
	ing1 := targetRecipe.Ingredient1
	ing2 := targetRecipe.Ingredient2

	// Set up the search
	queue := list.New()
	localVisited := make(map[string]bool)
	parent := make(map[string]Recipe)
	discovered := make(map[string]bool)
	depthMap := make(map[string]int)

	// Start with base elements
	for _, base := range baseElements {
		queue.PushBack(base)
		localVisited[base] = true
		discovered[base] = true
		depthMap[base] = 0
	}

	// Run BFS to discover all elements including target ingredients
	for queue.Len() > 0 && !shouldStop() {
		// Get next element to explore
		currentElement := queue.Remove(queue.Front()).(string)
		currentDepth := depthMap[currentElement]
		nodesVisitedCount.Add(1)

		// Check if we already found both ingredients of the target recipe
		if discovered[ing1] && discovered[ing2] {
			// Try to create the target element using the specific recipe
			if !discovered[targetElement] {
				// Record this specific recipe in the parent map
				parent[targetElement] = targetRecipe
				depthMap[targetElement] = max(depthMap[ing1], depthMap[ing2]) + 1
				discovered[targetElement] = true

				// Build and return the path
				return buildDiversePath(parent, targetElement, strategyVariant)
			}
		}

		// Get all elements to combine with
		combinableElements := make([]string, 0, len(discovered))
		for elem := range discovered {
			combinableElements = append(combinableElements, elem)
		}

		// Sort elements based on strategy
		sortElements(combinableElements, strategyVariant, depthMap, int(nodesVisitedCount.Load()))

		// Try combinations with other elements
		for _, otherElement := range combinableElements {
			// Generate unique pair key
			pairKey := getPairKey(currentElement, otherElement)

			// Skip if already visited
			if localVisited[pairKey] {
				continue
			}
			localVisited[pairKey] = true

			// Get recipes for this combination
			recipes := getRecipes(currentElement, otherElement)

			// Process each recipe result
			for _, recipe := range recipes {
				if shouldStop() {
					return []Recipe{}
				}

				result := recipe.Result
				resultDepth := currentDepth + 1

				// Update the parent and depth tracking
				_, alreadyFound := parent[result]

				// Give priority to the target recipe ingredients
				isPriorityElement := result == ing1 || result == ing2

				// Update parent if new or priority
				if !alreadyFound || isPriorityElement {
					parent[result] = recipe
					depthMap[result] = resultDepth
				}

				// Add to discovered elements
				wasNewDiscovery := !discovered[result]
				discovered[result] = true

				// Add to queue if not already visited or is priority
				if wasNewDiscovery || isPriorityElement {
					if !localVisited[result] || isPriorityElement {
						localVisited[result] = true

						// Give higher priority to target ingredients
						if isPriorityElement {
							queue.PushFront(result)
						} else {
							queue.PushBack(result)
						}
					}
				}
			}
		}
	}

	// If we get here, we didn't find a path using the specific combination
	return []Recipe{}
}

// Helper function to sort elements differently based on strategy
func sortElements(elements []string, strategyVariant int, depthMap map[string]int, seed int) {
	switch strategyVariant {
	case 0:
		// Alphabetical
		sort.Strings(elements)
	case 1:
		// Reverse alphabetical
		sort.Slice(elements, func(i, j int) bool {
			return elements[i] > elements[j]
		})
	case 2:
		// By depth (shallow first)
		sort.Slice(elements, func(i, j int) bool {
			depthI := depthMap[elements[i]]
			depthJ := depthMap[elements[j]]
			if depthI != depthJ {
				return depthI < depthJ
			}
			return elements[i] < elements[j]
		})
	case 3:
		// By depth (deep first)
		sort.Slice(elements, func(i, j int) bool {
			depthI := depthMap[elements[i]]
			depthJ := depthMap[elements[j]]
			if depthI != depthJ {
				return depthI > depthJ
			}
			return elements[i] < elements[j]
		})
	case 4:
		// Pseudo-random but deterministic ordering
		sort.Slice(elements, func(i, j int) bool {
			hashI := (seed*31 + len(elements[i])*43 + int(elements[i][0])) % 100
			hashJ := (seed*31 + len(elements[j])*43 + int(elements[j][0])) % 100
			if hashI != hashJ {
				return hashI < hashJ
			}
			return elements[i] < elements[j]
		})
	}
}

// buildDiversePath builds a valid path with worker-specific strategy for diversity
func buildDiversePath(parent map[string]Recipe, target string, workerID int) []Recipe {
	// Track elements needed for the path
	elementsNeeded := make(map[string]bool)

	// Build list of all required elements starting from target
	queue := list.New()
	queue.PushBack(target)
	processed := make(map[string]bool)
	processed[target] = true
	elementsNeeded[target] = true

	// Process all dependencies recursively
	for queue.Len() > 0 {
		current := queue.Remove(queue.Front()).(string)

		// Skip base elements
		if isBaseElement(current) {
			continue
		}

		// Get recipe for this element
		recipe, exists := parent[current]
		if !exists {
			return []Recipe{} // Invalid path
		}

		// Queue ingredients for processing
		for _, ingredient := range []string{recipe.Ingredient1, recipe.Ingredient2} {
			// Skip if already processed or is a base element
			if processed[ingredient] || isBaseElement(ingredient) {
				continue
			}

			processed[ingredient] = true
			elementsNeeded[ingredient] = true
			queue.PushBack(ingredient)
		}
	}

	// Build path in correct order
	var result []Recipe
	available := make(map[string]bool)

	// Start with base elements
	for _, base := range baseElements {
		available[base] = true
	}

	// Keep adding recipes until we have the target
	for !available[target] {
		// Find elements where ingredients are available
		candidates := make([]Recipe, 0)

		// Check each needed element
		for element := range elementsNeeded {
			// Skip if already available
			if available[element] {
				continue
			}

			// Get recipe
			recipe, exists := parent[element]
			if !exists {
				continue
			}

			// Check if ingredients are available
			if available[recipe.Ingredient1] && available[recipe.Ingredient2] {
				candidates = append(candidates, recipe)
			}
		}

		// If no valid candidates, path is invalid
		if len(candidates) == 0 {
			return []Recipe{}
		}

		// Sort candidates with slight worker-specific variations for diversity
		strategyVariant := workerID % 3
		sort.SliceStable(candidates, func(i, j int) bool {
			switch strategyVariant {
			case 0:
				// Strategy 0: Sort by result name (ascending)
				return candidates[i].Result < candidates[j].Result
			case 1:
				// Strategy 1: Sort by result name (descending)
				return candidates[i].Result > candidates[j].Result
			default:
				// Strategy 2: Sort by ingredient names
				ing1i := candidates[i].Ingredient1 + candidates[i].Ingredient2
				ing1j := candidates[j].Ingredient1 + candidates[j].Ingredient2
				return ing1i < ing1j
			}
		})

		// Pick first valid candidate
		recipe := candidates[0]
		result = append(result, recipe)
		available[recipe.Result] = true

		// If we have our target, we're done
		if available[target] {
			break
		}
	}

	return result
}

// Reset global caches - can be called to free memory or refresh state
func ResetCaches() {
	bfsPathCacheMutex.Lock()
	bfsPathCache = make(map[string][]Recipe)
	bfsPathCacheMutex.Unlock()
}
