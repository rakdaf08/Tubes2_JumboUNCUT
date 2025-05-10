import React from 'react';
import Tree from 'react-d3-tree';
import './SearchResults.css';

const API_BASE_URL = "http://localhost:8080";

const isBaseElement = (name) => {
    const baseElements = ["Air", "Earth", "Fire", "Water"];
    return baseElements.includes(name);
};

const buildElementNode = (elementName, pathRecipesMap, imageURLs, depth = 0) => {
     if (depth > 500) {
         console.error(`Max depth (${depth}) reached for ${elementName}. Stopping recursion.`);
         return {
              name: `${elementName} (MAX_DEPTH)`,
              attributes: { type: 'Error', originalName: elementName, depth: depth },
              children: []
         };
     }

     const node = {
         name: elementName,
         attributes: {
             type: isBaseElement(elementName) ? 'Base Element' : 'Element',
             imageUrl: imageURLs?.[elementName] || '',
             depth: depth
         },
         children: []
     };

     const recipeMakingThis = pathRecipesMap[elementName];

     if (recipeMakingThis) {
         const recipeNode = buildRecipeNode(recipeMakingThis, pathRecipesMap, imageURLs, depth + 1);
         node.children.push(recipeNode);
     }

     return node;
};

const buildRecipeNode = (recipe, pathRecipesMap, imageURLs, depth = 0) => {
   if (depth > 500) {
        console.error(`Max depth (${depth}) reached for recipe ${recipe.ingredient1} + ${recipe.ingredient2} => ${recipe.result}. Stopping recursion.`);
         return {
            name: `(${recipe.ingredient1} + ${recipe.ingredient2}) (MAX_DEPTH)`,
            attributes: { type: 'ErrorRecipe', result: recipe.result, depth: depth },
            children: []
        };
   }

  const node = {
      name: `${recipe.ingredient1} + ${recipe.ingredient2}`,
      attributes: {
          type: 'Recipe',
          result: recipe.result,
          ingredient1: recipe.ingredient1,
          ingredient2: recipe.ingredient2,
          depth: depth
      },
      children: []
  };

  const ingredient1Node = buildElementNode(recipe.ingredient1, pathRecipesMap, imageURLs, depth + 1);
  const ingredient2Node = buildElementNode(recipe.ingredient2, pathRecipesMap, imageURLs, depth + 1);

  node.children.push(ingredient1Node);
  node.children.push(ingredient2Node);

  return node;
};

const buildTreeData = (path, targetElement, imageURLs) => {
  if (!path || path.length === 0 || isBaseElement(targetElement)) {
     const isBase = isBaseElement(targetElement);
     const rootNode = {
         name: targetElement,
         attributes: {
             type: isBase ? 'Base Element' : 'Target Element',
             imageUrl: imageURLs?.[targetElement] || '',
             depth: 0
         },
         children: []
     };
     return [rootNode];
  }

  const pathRecipesMap = {};
  path.forEach(recipe => {
      pathRecipesMap[recipe.result] = recipe;
  });

  const rootNode = buildElementNode(targetElement, pathRecipesMap, imageURLs, 0);

  return [rootNode];
};

const renderNodeWithImage = ({ nodeDatum, toggleNode }) => {
    const isRecipeNode = nodeDatum.attributes?.type === 'Recipe';
    const elementNameToFetch = nodeDatum.name || '';
    const imageUrlFromBackendProxy = elementNameToFetch && nodeDatum.attributes?.type !== 'Recipe'
    ? `${API_BASE_URL}/api/image?elementName=${encodeURIComponent(elementNameToFetch)}`
    : '';

    const imageSize = 50;
    const textYOffset = imageSize / 2 + 10;
    const textXOffset = 0;

    return (
      <g onClick={toggleNode}>
        {!isRecipeNode && (
            <>
                {imageUrlFromBackendProxy && (
                   <image
                      x={-imageSize / 2}
                      y={-imageSize / 2}
                      width={imageSize}
                      height={imageSize}
                      href={imageUrlFromBackendProxy}
                      onError={(e) => {
                           e.target.style.display = 'none';
                       }}
                       title={`Elemen: ${nodeDatum.name}`}
                   />
                )}

                <text
                  strokeWidth="0.5"
                  x={textXOffset}
                  y={textYOffset}
                  textAnchor="middle"
                  alignmentBaseline="middle"
                  className="element-node-text"
                >
                  {nodeDatum.name}
                </text>
            </>
        )}

        {isRecipeNode && (
             <>
                 <rect className="recipe-node-rect"/>
                 <text
                   strokeWidth="0.5"
                   x={0}
                   y={5}
                   textAnchor="middle"
                   alignmentBaseline="middle"
                   className="recipe-node-text"
                 >
                   {nodeDatum.name}
                 </text>
             </>
        )}
      </g>
    );
};


function SearchResults({ results, isLoading, error }) {
  if (isLoading) {
    return <div className="loading-message">Loading...</div>;
  }

  if (error) {
    const errorMessage = (typeof error === 'object' && error !== null && error.message) ? error.message : String(error);
    return <div className="error-message">Error: {errorMessage}</div>;
  }

  if (!results || typeof results !== 'object') {
    return <p className="initial-message">Masukkan elemen target dan klik cari.</p>;
  }

   const handleImageError = (e) => {
       e.target.style.display = 'none';
       const altText = e.target.alt || 'image';
       const errorSpan = document.createElement('span');
       errorSpan.textContent = `(${altText} error)`;
       errorSpan.style.fontSize = '0.8em';
       errorSpan.style.color = '#dc3545';
       if (e.target.parentNode) {
           e.target.parentNode.insertBefore(errorSpan, e.target.nextSibling);
       }
   };

  const renderStep = (step, stepIndex, pathIndex = null) => {
    if (!step || typeof step !== 'object') {
        return <li key={`invalid-step-${pathIndex}-${stepIndex}`} className="recipe-step-item invalid">Data langkah tidak valid</li>;
    }
    const key = `${pathIndex !== null ? pathIndex + '-' : ''}${stepIndex}`;

    const imageUrlPath1 = results.imageURLs?.[step.ingredient1];
    const imageUrlPath2 = results.imageURLs?.[step.ingredient2];
    const imageUrlPathResult = results.imageURLs?.[step.result];

    const imageUrl1 = imageUrlPath1 ? `${API_BASE_URL}${imageUrlPath1}` : '';
    const imageUrl2 = imageUrlPath2 ? `${API_BASE_URL}${imageUrlPath2}` : '';
    const imageUrlResult = imageUrlPathResult ? `${API_BASE_URL}${imageUrlPathResult}` : '';

    return (
      <li key={key} className="recipe-step-item">
        {imageUrl1 ? <img src={imageUrl1} alt={step.ingredient1 || 'ingredient1'} className="recipe-step-image" onError={handleImageError}/> : null}
        {step.ingredient1 || '?'}
        <span className="recipe-step-separator">+</span>
         {imageUrl2 ? <img src={imageUrl2} alt={step.ingredient2 || 'ingredient2'} className="recipe-step-image" onError={handleImageError}/> : null}
        {step.ingredient2 || '?'}
        <span className="recipe-step-separator">{' => '}</span>
         {imageUrlResult ? <img src={imageUrlResult} alt={step.result || 'result'} className="recipe-step-image" onError={handleImageError}/> : null}
        <strong className="recipe-step-result">{step.result || '?'}</strong>
      </li>
    );
  };

  const renderPath = (path, pathIndex = null) => (
    Array.isArray(path) ? (
        <ol key={pathIndex !== null ? `path-${pathIndex}` : 'single-path-list'} className="recipe-path-list">
          {path.map((step, stepIndex) => renderStep(step, stepIndex, pathIndex))}
        </ol>
    ) : <p key={pathIndex !== null ? `invalid-path-list-${pathIndex}` : 'invalid-single-path-list'} className="invalid-path-message">Data jalur tidak valid (bukan array).</p>
  );

  let treeDataForRendering = [];

  if (results.pathFound === true) {
      if (!isBaseElement(results.searchTarget)) {
          if (results.mode === 'shortest' && results.path) {
               treeDataForRendering = buildTreeData(results.path, results.searchTarget, results.imageURLs);
          } else if (results.mode === 'multiple' && results.paths && results.paths.length > 0) {
               treeDataForRendering = results.paths.map(path => buildTreeData(path, results.searchTarget, results.imageURLs)[0]);
          }
      } else {
           treeDataForRendering = buildTreeData([], results.searchTarget, results.imageURLs);
      }
  }

  return (
    <div className="search-results-container">
      <h2 className="results-title">
        Hasil Pencarian untuk: <strong className="target-element">{results.searchTarget || 'N/A'}</strong>
        <span className="search-info">
          ({results.algorithm?.toUpperCase() || 'N/A'}/{results.mode || 'N/A'}
          {results.mode === 'multiple' && ` - Max: ${results.maxRecipes || 'N/A'}`})
        </span>
      </h2>

      {results.pathFound === true && (
         <div className="search-stats">
             <span>Node Dikunjungi: <strong>{results.nodesVisited !== undefined && results.nodesVisited !== -1 ? results.nodesVisited : 'N/A'}</strong></span>
             <span>Durasi: <strong>{results.durationMillis ?? 'N/A'} ms</strong></span>
         </div>
      )}

       {results.pathFound === true ? (
           <>
           {results.pathFound === true && isBaseElement(results.searchTarget) && (
               ((results.mode === 'shortest' && (!results.path || results.path.length === 0)) ||
                (results.mode === 'multiple' && (!results.paths || results.paths.length === 0 || (results.paths.length > 0 && results.paths.every(path => path.length === 0)))))
           ) && (
                <div className="base-element-message">
                     (Target adalah elemen dasar, tidak ada resep pembuat.)
                </div>
           )}

           {
               (results.mode === 'shortest' && results.path && results.path.length > 0) ?
               [results.path].map((path, index) => (
                   <div key={`path-block-${index}`} className="path-block">
                       <div className="path-text-section">
                           <h4 className="path-title">
                             Jalur {index + 1} (Langkah: {path.length})
                           </h4>
                           {renderPath(path, index)}
                       </div>

                       {treeDataForRendering.length > index && treeDataForRendering[index] && (
                           <div className="path-visualization-section">
                                <h4 className="visualization-title">Visualisasi Jalur {index + 1}:</h4>
                                <div id={`treeWrapper-${index}`} className="tree-wrapper">
                                    <Tree
                                        data={[treeDataForRendering[index]]}
                                        orientation="vertical"
                                        translate={{ x: 250, y: 50 }}
                                        renderCustomNodeElement={renderNodeWithImage}
                                        zoomable={true}
                                        draggable={true}
                                        nodeSize={{ x: 120, y: 100 }}
                                        separation={{ siblings: 1, nonSiblings: 1 }}
                                    />
                                </div>
                           </div>
                       )}
                   </div>
               ))
               :
               (results.mode === 'multiple' && results.paths && results.paths.length > 0) ?
               results.paths.map((path, index) => (
                   <div key={`path-block-${index}`} className="path-block">
                       <div className="path-text-section">
                           <h4 className="path-title">
                             Jalur {index + 1} (Langkah: {path.length})
                           </h4>
                           {renderPath(path, index)}
                       </div>

                       {treeDataForRendering.length > index && treeDataForRendering[index] && (
                           <div className="path-visualization-section">
                                <h4 className="visualization-title">Visualisasi Jalur {index + 1}:</h4>
                                <div id={`treeWrapper-${index}`} className="tree-wrapper">
                                     <Tree
                                         data={[treeDataForRendering[index]]}
                                         orientation="vertical"
                                         translate={{ x: 250, y: 50 }}
                                         renderCustomNodeElement={renderNodeWithImage}
                                         zoomable={true}
                                         draggable={true}
                                         nodeSize={{ x: 120, y: 100 }}
                                         separation={{ siblings: 1, nonSiblings: 1 }}
                                     />
                                 </div>
                           </div>
                       )}
                   </div>
               ))
               :
               null
           }
           </>
       ) : (
            <div className="path-not-found-message">
                 Jalur tidak ditemukan untuk elemen "{results.searchTarget || 'N/A'}". {results.error ? `(${results.error})` : ''}
            </div>
       )}

      {results.imageURLs && typeof results.imageURLs === 'object' && Object.keys(results.imageURLs).length > 0 && (
           <details className="image-urls-details">
              <summary className="image-urls-summary">Tampilkan/Sembunyikan URL Gambar Terkait</summary>
              <ul className="image-urls-list">
                {Object.entries(results.imageURLs).map(([name, url]) => (
                  <li key={name} className="image-urls-item">
                    {url ? (
                         <img
                           src={`${API_BASE_URL}${url}`}
                           alt={name || '?'}
                           className="image-urls-image"
                           onError={handleImageError}
                         />
                    ) : null}
                    <span className="image-urls-name">{name || 'N/A'}:</span>
                    <a href={`${API_BASE_URL}${url}` || '#'} target="_blank" rel="noopener noreferrer" className="image-urls-link">{url || 'N/A'}</a>
                  </li>
                ))}
              </ul>
           </details>
        )}
    </div>
  );
}

export default SearchResults;
