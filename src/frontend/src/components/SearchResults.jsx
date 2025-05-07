import React from 'react';
import Tree from 'react-d3-tree';

const isBaseElement = (name) => {
    const baseElements = ["Air", "Earth", "Fire", "Water"];
    return baseElements.includes(name);
};

const buildElementNode = (elementName, pathRecipesMap, imageURLs) => {
     const node = {
         name: elementName,
         attributes: {
             type: isBaseElement(elementName) ? 'Base Element' : 'Element',
             imageUrl: imageURLs?.[elementName] || ''
         },
         children: []
     };

     const recipeMakingThis = pathRecipesMap[elementName];

     if (recipeMakingThis) {
         const recipeNode = buildRecipeNode(recipeMakingThis, pathRecipesMap, imageURLs);
         node.children.push(recipeNode);
     }

     return node;
};

const buildRecipeNode = (recipe, pathRecipesMap, imageURLs) => {
    const node = {
        name: `${recipe.ingredient1} + ${recipe.ingredient2}`,
        attributes: {
            type: 'Recipe',
            result: recipe.result,
            ingredient1: recipe.ingredient1,
            ingredient2: recipe.ingredient2,
        },
        children: []
    };

    const ingredient1Node = buildElementNode(recipe.ingredient1, pathRecipesMap, imageURLs);
    const ingredient2Node = buildElementNode(recipe.ingredient2, pathRecipesMap, imageURLs);

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
             imageUrl: imageURLs?.[targetElement] || ''
         },
         children: []
     };
     return [rootNode];
  }

  const pathRecipesMap = {};
  path.forEach(recipe => {
      pathRecipesMap[recipe.result] = recipe;
  });

  const rootNode = buildElementNode(targetElement, pathRecipesMap, imageURLs);

  return [rootNode];
};

const renderNodeWithImage = ({ nodeDatum, toggleNode }) => {
    const isRecipeNode = nodeDatum.attributes?.type === 'Recipe';

    return (
      <g onClick={toggleNode}>
        {!isRecipeNode && (
            <>
                <circle r={20} fill="#4682B4" stroke="#000" strokeWidth="1.5" />

                {nodeDatum.attributes?.imageUrl && (
                   <image
                      x={-15}
                      y={-35}
                      width={30}
                      height={30}
                      href={nodeDatum.attributes.imageUrl}
                      onError={(e) => {
                           console.warn(`Failed to load image for ${nodeDatum.name}: ${e.target.href}`);
                           e.target.style.display = 'none';
                       }}
                       title={`Elemen: ${nodeDatum.name}`}
                   />
                )}

                <text
                  strokeWidth="0.5"
                  x={0}
                  y={10}
                  textAnchor="middle"
                  alignmentBaseline="middle"
                  style={{
                      fontSize: '14px',
                      fill: '#fff',
                      fontWeight: 'bold',
                      pointerEvents: 'none',
                  }}
                >
                  {nodeDatum.name}
                </text>
            </>
        )}

        {isRecipeNode && (
             <>
                 <rect
                    x={-60}
                    y={-15}
                    width={120}
                    height={30}
                    rx={5}
                    ry={5}
                    fill="#e9ecef"
                    stroke="#adb5bd"
                    strokeWidth="1"
                 />
                 <text
                   strokeWidth="0.5"
                   x={0}
                   y={5}
                   textAnchor="middle"
                   alignmentBaseline="middle"
                   style={{
                       fontSize: '12px',
                       fill: '#495057',
                       pointerEvents: 'none',
                   }}
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
    return <div style={{ textAlign: 'center', padding: '20px', fontSize: '18px', color: '#333' }}>Loading...</div>;
  }

  if (error) {
    const errorMessage = (typeof error === 'object' && error !== null && error.message) ? error.message : String(error);
    return <div style={{ color: '#721c24', border: '1px solid #f5c6cb', background: '#f8d7da', padding: '15px', borderRadius: '5px', marginTop: '20px' }}>Error: {errorMessage}</div>;
  }

  if (!results || typeof results !== 'object') {
    return <p style={{ textAlign: 'center', marginTop: '20px', color: '#6c757d' }}>Masukkan elemen target dan klik cari.</p>;
  }

  const handleImageError = (e) => {
      console.warn(`Image failed to load: ${e.target.src}`);
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
        return <li key={`${pathIndex}-${stepIndex}`} style={{color: 'red'}}>Data langkah tidak valid</li>;
    }
    const key = `${pathIndex}-${stepIndex}`;
    const imageUrl1 = results.imageURLs?.[step.ingredient1];
    const imageUrl2 = results.imageURLs?.[step.ingredient2];
    const imageUrlResult = results.imageURLs?.[step.result];

    const imgStyle = {
        height: '20px',
        width: '20px',
        verticalAlign: 'middle',
        margin: '0 3px',
        border: '1px solid #eee',
        objectFit: 'contain'
    };

    return (
      <li key={key} style={{ marginBottom: '8px', lineHeight: '1.5', color: '#212529' }}>
        {imageUrl1 ? <img src={imageUrl1} alt={step.ingredient1 || 'ingredient1'} style={imgStyle} onError={handleImageError}/> : null}
        {step.ingredient1 || '?'}
        <span style={{ margin: '0 5px' }}>+</span>
        {imageUrl2 ? <img src={imageUrl2} alt={step.ingredient2 || 'ingredient2'} style={imgStyle} onError={handleImageError}/> : null}
        {step.ingredient2 || '?'}
        <span style={{ margin: '0 5px' }}>{' => '}</span>
        {imageUrlResult ? <img src={imageUrlResult} alt={step.result || 'result'} style={imgStyle} onError={handleImageError}/> : null}
        <strong style={{ color: '#000' }}>{step.result || '?'}</strong>
      </li>
    );
  };

  const renderPath = (path, pathIndex = null) => (
    Array.isArray(path) ? (
        <ol key={pathIndex} style={{ paddingLeft: '20px', marginTop: '5px' }}>
          {path.map((step, index) => renderStep(step, index, pathIndex))}
        </ol>
    ) : <p style={{ color: '#dc3545' }}>Data jalur tidak valid (bukan array).</p>
  );


  let treeDataForRendering = [];

  if (results.pathFound === true) {
      if (results.mode === 'shortest' && results.path) {
           treeDataForRendering = buildTreeData(results.path, results.searchTarget, results.imageURLs);
      } else if (results.mode === 'multiple' && results.paths && results.paths.length > 0) {
           treeDataForRendering = results.paths.map(path => buildTreeData(path, results.searchTarget, results.imageURLs)[0]);
      } else if (['Air', 'Earth', 'Fire', 'Water'].includes(results.searchTarget)) {
         treeDataForRendering = buildTreeData([], results.searchTarget, results.imageURLs);
      }
  }


  return (
    <div style={{ marginTop: '30px', borderTop: '2px solid #0d6efd', paddingTop: '20px', background: '#ffffff', padding: '20px', borderRadius: '8px', boxShadow: '0 2px 4px rgba(0,0,0,0.1)', color: '#212529' }}>

      <h2 style={{ borderBottom: '1px solid #dee2e6', paddingBottom: '10px', marginBottom: '20px', color: '#343a40' }}>
        Hasil Pencarian untuk: <strong style={{ color: '#0d6efd' }}>{results.searchTarget || 'N/A'}</strong>
        <span style={{ fontSize: '0.9em', color: '#6c757d', marginLeft: '10px' }}>
          ({results.algorithm?.toUpperCase() || 'N/A'}/{results.mode || 'N/A'}
          {results.mode === 'multiple' && ` - Max: ${results.maxRecipes || 'N/A'}`})
        </span>
      </h2>

      <div style={{ marginBottom: '20px', display: 'flex', justifyContent: 'space-between', flexWrap: 'wrap', gap: '10px', fontSize: '0.95em', color: '#495057' }}>
          <span>Node Dikunjungi: <strong>{results.nodesVisited !== undefined && results.nodesVisited !== -1 ? results.nodesVisited : 'N/A'}</strong></span>
          <span>Durasi: <strong>{results.durationMillis ?? 'N/A'} ms</strong></span>
      </div>


       {/* AREA UNTUK MENAMPILKAN HASIL TEKS */}
       {results.pathFound === true && (
           <>
           {results.mode === 'shortest' && results.path && results.path.length > 0 && (
                <div style={{ background: '#f8f9fa', padding: '15px', borderRadius: '5px', border: '1px solid #dee2e6', marginBottom: '20px' }}>
                   <h3 style={{ marginTop: '0', marginBottom: '10px', color: '#495057' }}>Jalur Resep Terpendek (Teks):</h3>
                   <p style={{ marginBottom: '10px', fontSize: '0.9em', color: '#6c757d' }}>Jumlah Langkah: {results.path.length}</p>
                   {renderPath(results.path)}
               </div>
           )}

           {results.mode === 'multiple' && results.paths && results.paths.length > 0 && (
              <div style={{ marginBottom: '20px' }}>
                 <h3 style={{ marginBottom: '15px', color: '#495057' }}>Jalur Resep Ditemukan (Teks):</h3>
                 {results.paths.map((path, index) => (
                     <div key={`text-path-${index}`} style={{ marginBottom: '15px', border: '1px solid #eee', borderRadius: '5px', padding: '15px' }}>
                         <h4 style={{ marginTop: '0', marginBottom: '10px', borderBottom: '1px solid #eee', paddingBottom: '5px', color: '#495057' }}>
                           Jalur {index + 1} (Langkah: {path.length})
                         </h4>
                         {renderPath(path, index)}
                     </div>
                 ))}
              </div>
           )}

           {results.pathFound === true && (
              (results.mode === 'shortest' && (!results.path || results.path.length === 0)) ||
              (results.mode === 'multiple' && (!results.paths || results.paths.length === 0 || (results.paths.length > 0 && results.paths.every(path => path.length === 0))))
           ) && isBaseElement(results.searchTarget) && (
                <div style={{ color: '#6c757d', border: '1px solid #ced4da', background: '#e9ecef', padding: '15px', borderRadius: '5px', marginTop: '20px', marginBottom: '20px' }}>
                     (Target adalah elemen dasar, tidak ada resep pembuat.)
                </div>
           )}
           </>
       )}


      {/* AREA UNTUK MENAMPILKAN VISUALISASI POHON */}
      {results.pathFound === true && treeDataForRendering.length > 0 && !isBaseElement(results.searchTarget) && (
          <>
              <h3 style={{ marginBottom: '15px', color: '#495057' }}>Visualisasi Pohon Resep:</h3>

             {results.mode === 'shortest' && (
                 <div style={{ background: '#f8f9fa', padding: '15px', borderRadius: '5px', border: '1px solid #dee2e6' }}>
                    <h4 style={{ marginTop: '0', marginBottom: '10px', color: '#495057' }}>Jalur Terpendek:</h4>
                    <div id="treeWrapper-shortest" style={{ width: '100%', height: '500px', border: '1px solid #ccc', overflow: 'auto' }}>
                        <Tree
                            data={treeDataForRendering}
                            orientation="vertical"
                            translate={{ x: 250, y: 50 }}
                            renderCustomNodeElement={renderNodeWithImage}
                            zoomable={true}
                            draggable={true}
                        />
                    </div>
                 </div>
              )}

              {results.mode === 'multiple' && results.paths && results.paths.length > 0 && (
                 <div>
                    {results.paths.map((path, index) => {
                        const singleTreeData = buildTreeData(path, results.searchTarget, results.imageURLs);
                        if (singleTreeData.length === 0) return null;

                        return (
                             <div key={`tree-path-${index}`} style={{ marginBottom: '20px', border: '1px solid #dee2e6', borderRadius: '5px', padding: '15px', background: '#ffffff' }}>
                                 <h4 style={{ marginTop: '0', marginBottom: '10px', borderBottom: '1px solid #eee', paddingBottom: '5px', color: '#495057' }}>
                                   Jalur {index + 1}:
                                 </h4>
                                 <div id={`treeWrapper-multiple-${index}`} style={{ width: '100%', height: '500px', border: '1px solid #ccc', overflow: 'auto' }}>
                                     <Tree
                                         data={singleTreeData}
                                         orientation="vertical"
                                         translate={{ x: 250, y: 50 }}
                                         renderCustomNodeElement={renderNodeWithImage}
                                         zoomable={true}
                                         draggable={true}
                                     />
                                 </div>
                             </div>
                        );
                    })}
                 </div>
              )}
          </>
      )}

        {results.pathFound === true && treeDataForRendering.length === 0 && isBaseElement(results.searchTarget) && (
               <div style={{ color: '#6c757d', border: '1px solid #ced4da', background: '#e9ecef', padding: '15px', borderRadius: '5px', marginTop: '20px' }}>
                    (Target adalah elemen dasar, tidak ada visualisasi pohon resep.)
               </div>
           )}
         {results.pathFound === true && treeDataForRendering.length === 0 && !isBaseElement(results.searchTarget) && (
             <div style={{ color: '#856404', border: '1px solid #ffeeba', background: '#fff3cd', padding: '15px', borderRadius: '5px', marginTop: '20px' }}>
                 Jalur resep ditemukan, tetapi visualisasi pohon tidak tersedia atau jalur kosong. Mohon periksa data resep dari backend.
             </div>
         )}


      {/* AREA UNTUK MENAMPILKAN URL GAMBAR */}
      {results.imageURLs && typeof results.imageURLs === 'object' && Object.keys(results.imageURLs).length > 0 && (
           <details style={{ marginTop: '20px', border: '1px solid #ddd', borderRadius: '5px', background: '#ffffff' }}>
              <summary style={{ padding: '10px', cursor: 'pointer', fontWeight: 'bold', color: '#495057' }}>Tampilkan/Sembunyikan URL Gambar Terkait</summary>
              <ul style={{ listStyle: 'none', padding: '15px', margin: '0', maxHeight: '150px', overflowY: 'auto' }}>
                {Object.entries(results.imageURLs).map(([name, url]) => (
                  <li key={name} style={{ marginBottom: '8px', display: 'flex', alignItems: 'center', color: '#212529' }}>
                    <img
                      src={url || ''}
                      alt={name || '?'}
                      style={{ height: '24px', width: '24px', verticalAlign: 'middle', marginRight: '8px', border: '1px solid #eee', objectFit: 'contain' }}
                       onError={(e) => { e.target.style.display = 'none'; }}
                    />
                    <span style={{ fontWeight: '500', marginRight: '5px' }}>{name || 'N/A'}:</span>
                    <a href={url || '#'} target="_blank" rel="noopener noreferrer" style={{ fontSize: '0.85em', color: '#0d6efd', wordBreak: 'break-all' }}>{url || 'N/A'}</a>
                  </li>
                ))}
              </ul>
           </details>
        )}


    </div>
  );
}

export default SearchResults;