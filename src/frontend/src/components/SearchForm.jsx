import React, { useState } from 'react';
import './SearchForm.css';

function SearchForm({ onSearchSubmit, isLoading }) {
  const [target, setTarget] = useState('');
  const [algo, setAlgo] = useState('bfs');
  const [mode, setMode] = useState('shortest');
  const [maxRecipes, setMaxRecipes] = useState(1);

  const handleSubmit = (event) => {
    event.preventDefault();
    if (!target) {
      alert('Masukkan elemen target!');
      return;
    }
    if (mode === 'multiple' && (!maxRecipes || maxRecipes <= 0)) {
        alert('Masukkan jumlah resep minimal 1 untuk mode multiple!');
        return;
    }

    const searchParams = { target, algo, mode };
    if (mode === 'multiple') {
        searchParams.max = maxRecipes;
    }
    onSearchSubmit(searchParams);
  };

  return (
    <form onSubmit={handleSubmit} className="search-form-container">
      <div>
        <label htmlFor="targetElement" className="form-label">Elemen Target:</label>
        <input
          type="text"
          id="targetElement"
          value={target}
          onChange={(e) => setTarget(e.target.value)}
          placeholder="Contoh: Mud, Human, ..."
          required
          className="form-input"
        />
      </div>

      <div className="form-options-group">
        <p className="options-title">Algoritma:</p>
        <div className="radio-group">
            <label className="radio-label">
              <input
                type="radio"
                value="bfs"
                checked={algo === 'bfs'}
                onChange={(e) => setAlgo(e.target.value)}
                className="radio-input"
              /> BFS (Shortest Path)
            </label>
            <label className="radio-label">
              <input
                type="radio"
                value="dfs"
                checked={algo === 'dfs'}
                onChange={(e) => setAlgo(e.target.value)}
                 className="radio-input"
              /> DFS (A Path / Multiple)
            </label>
        </div>
      </div>

      <div className="form-options-group">
        <p className="options-title">Mode:</p>
        <div className="radio-group">
            <label className="radio-label">
              <input
                type="radio"
                value="shortest"
                checked={mode === 'shortest'}
                onChange={(e) => setMode(e.target.value)}
                 className="radio-input"
              /> Shortest
            </label>
            <label className="radio-label">
              <input
                type="radio"
                value="multiple"
                checked={mode === 'multiple'}
                onChange={(e) => setMode(e.target.value)}
                 className="radio-input"
              /> Multiple
            </label>
        </div>
         {mode === 'multiple' && (
            <div className="max-recipes-group">
                 <label htmlFor="maxRecipes" className="max-recipes-label">Jumlah Resep:</label>
                 <input
                    type="number"
                    id="maxRecipes"
                    value={maxRecipes}
                    onChange={(e) => setMaxRecipes(parseInt(e.target.value, 10) || 1)}
                    min="1"
                    className="max-recipes-input"
                 />
            </div>
         )}
      </div>

      <button
        type="submit"
        disabled={isLoading}
        className="submit-button"
      >
        {isLoading ? 'Mencari...' : 'Cari Resep'}
      </button>
    </form>
  );
}

export default SearchForm;
