import React, { useState, useEffect } from 'react';
import SearchForm from '../components/SearchForm';
import SearchResults from '../components/SearchResults';
import { findRecipes } from '../api/searchService';
import './SearchPage.css';

function SearchPage() {
  const [searchResults, setSearchResults] = useState(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);
  const [currentParams, setCurrentParams] = useState(null);

  const handleSearch = async (searchParams) => {
    setCurrentParams(searchParams);
    setIsLoading(true);
    setError(null);
    setSearchResults(null);

    try {
      const { target, algo, mode, max } = searchParams;
      const data = await findRecipes(target, algo, mode, max);
      setSearchResults(data);
    } catch (err) {
      setError(err.message || 'Terjadi kesalahan saat mencari resep.');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="search-page-container">
      <h1 className="page-title">Pencari Resep Little Alchemy 2</h1>
      <SearchForm onSearchSubmit={handleSearch} isLoading={isLoading} />
      <SearchResults results={searchResults} isLoading={isLoading} error={error} />
    </div>
  );
}

export default SearchPage;
