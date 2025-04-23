// src/frontend/src/components/SearchResults.jsx
import React from 'react';

function SearchResults({ results, isLoading, error }) {
  if (isLoading) {
    return <p>Loading...</p>;
  }

  if (error) {
    return <p style={{ color: 'red' }}>Error: {error}</p>;
  }

  if (!results) {
    return <p>Masukkan elemen target dan klik cari.</p>;
  }

  // Tampilan hasil awal (sebelum visualisasi pohon)
  return (
    <div>
      <h2>Hasil Pencarian untuk: {results.searchTarget} ({results.algorithm}/{results.mode})</h2>
      {results.pathFound ? (
        <>
          <p>Node Dikunjungi: {results.nodesVisited}</p>
          <p>Durasi: {results.durationMillis} ms</p>
          <p>Jumlah Langkah Resep: {results.path ? results.path.length : 0}</p>
          <h3>Jalur Resep:</h3>
          {results.path && results.path.length > 0 ? (
            <ol>
              {results.path.map((step, index) => (
                <li key={index}>
                  {step.ingredient1} + {step.ingredient2} {'=>'} {step.result}
                </li>
              ))}
            </ol>
          ) : (
             <p>(Target adalah elemen dasar)</p>
          )}
          <h3>URL Gambar Terkait:</h3>
          {results.imageURLs && Object.keys(results.imageURLs).length > 0 ? (
            <ul>
              {Object.entries(results.imageURLs).map(([name, url]) => (
                <li key={name}>
                  {name}: <a href={url} target="_blank" rel="noopener noreferrer">{url}</a>
                  <img src={url} alt={name} style={{height: '20px', verticalAlign: 'middle', marginLeft: '5px'}}/>
                </li>
              ))}
            </ul>
          ) : (
            <p>(Tidak ada URL gambar ditemukan untuk path ini)</p>
          )}
        </>
      ) : (
        <p>Jalur tidak ditemukan.</p>
      )}
      {/* Tampilkan raw JSON untuk debug (opsional) */}
      {/* <pre>{JSON.stringify(results, null, 2)}</pre> */}
    </div>
  );
}

export default SearchResults;