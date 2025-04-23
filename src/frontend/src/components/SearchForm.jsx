// src/frontend/src/components/SearchForm.jsx
import React, { useState } from 'react';

// Menerima prop 'onSearchSubmit' dari parent (SearchPage)
function SearchForm({ onSearchSubmit, isLoading }) {
  const [target, setTarget] = useState('');
  const [algo, setAlgo] = useState('bfs'); // Default BFS
  const [mode, setMode] = useState('shortest'); // Default shortest

  const handleSubmit = (event) => {
    event.preventDefault(); // Mencegah refresh halaman standar form HTML
    if (!target) {
      alert('Masukkan elemen target!');
      return;
    }
    // Panggil fungsi yang di-pass dari parent dengan data dari state
    onSearchSubmit({ target, algo, mode });
  };

  return (
    <form onSubmit={handleSubmit}>
      <div>
        <label htmlFor="targetElement">Elemen Target:</label>
        <input
          type="text"
          id="targetElement"
          value={target}
          onChange={(e) => setTarget(e.target.value)}
          placeholder="Contoh: Mud, Human, ..."
          required
        />
      </div>

      <div>
        <p>Algoritma:</p>
        <label>
          <input
            type="radio"
            value="bfs"
            checked={algo === 'bfs'}
            onChange={(e) => setAlgo(e.target.value)}
          /> BFS (Shortest Path)
        </label>
        <label>
          <input
            type="radio"
            value="dfs"
            checked={algo === 'dfs'}
            onChange={(e) => setAlgo(e.target.value)}
          /> DFS (A Path)
        </label>
      </div>

      <div>
        <p>Mode:</p>
         <label>
          <input
            type="radio"
            value="shortest"
            checked={mode === 'shortest'}
            onChange={(e) => setMode(e.target.value)}
          /> Shortest
        </label>
         <label>
          <input
            type="radio"
            value="multiple"
            checked={mode === 'multiple'}
            onChange={(e) => setMode(e.target.value)}
          /> Multiple (sementara pakai DFS)
        </label>
         {/* TODO: Tambah input 'max' jika mode == 'multiple' */}
      </div>

      <button type="submit" disabled={isLoading}>
        {isLoading ? 'Mencari...' : 'Cari Resep'}
      </button>
    </form>
  );
}

export default SearchForm;