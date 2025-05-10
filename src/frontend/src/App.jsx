// src/frontend/src/App.jsx
import React, { useState, useEffect } from 'react';
import './App.css';
import SearchForm from './components/SearchForm'; // Pastikan ini sudah di-uncomment
import SearchResults from './components/SearchResults';
import { findRecipes } from './api/searchService';

import jumboKocokLogo from './assets/Jumbo_Kocok.png';
import clickToStartImage from './assets/ClickToStart.png';
import sansImage from './assets/sans.jpg';

// Definisikan di luar komponen karena tidak berubah
const FULL_DIALOG_TEXT = "* Please enter the recipe you are looking for";
const TYPING_DELAY_START_MS = 450;
const TYPING_SPEED_MS = 70;

function App() {
  const [appState, setAppState] = useState('splash');
  const [logoPositionClass, setLogoPositionClass] = useState('center');
  const [resultsViewActive, setResultsViewActive] = useState(false);

  const [searchResultsData, setSearchResultsData] = useState(null); // Menggunakan nama yang berbeda untuk data hasil
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);
  // eslint-disable-next-line no-unused-vars
  const [currentParams, setCurrentParams] = useState(null);

  const [dialogText, setDialogText] = useState('');
  const [isDialogTyping, setIsDialogTyping] = useState(false);

  const handleInitialClick = () => {
    if (appState === 'splash') {
      setLogoPositionClass('topLeft');
      setAppState('logoMoving');
    }
  };

  // useEffect untuk transisi state utama aplikasi
  useEffect(() => {
    let timer;
    if (appState === 'logoMoving') {
      timer = setTimeout(() => {
        setAppState('sansAppearing');
      }, 1000);
    } else if (appState === 'sansAppearing') {
      const dialogBoxAnimationDuration = 1000;
      const sansAppearanceTime = 700;
      const textLength = FULL_DIALOG_TEXT.length;
      const totalTypingTimeMs = TYPING_DELAY_START_MS + (textLength * TYPING_SPEED_MS);
      const searchFormTransitionDelay = 1200; // ms (dari CSS .search-form-wrapper transition delay)

      const timeUntilContentReady = Math.max(dialogBoxAnimationDuration, sansAppearanceTime, totalTypingTimeMs, searchFormTransitionDelay) + 200;

      timer = setTimeout(() => {
        setAppState('contentReady');
      }, timeUntilContentReady);
    }
    return () => clearTimeout(timer);
  }, [appState]);

  // useEffect untuk animasi ketik dialog
  useEffect(() => {
    let typingTimerId;
    let charIndex = 0;

    if (appState === 'sansAppearing') {
      setDialogText('');
      setIsDialogTyping(true);

      const typeCharacter = () => {
        if (charIndex < FULL_DIALOG_TEXT.length) {
          setDialogText(prev => prev + FULL_DIALOG_TEXT.charAt(charIndex));
          charIndex++;
          typingTimerId = setTimeout(typeCharacter, TYPING_SPEED_MS);
        } else {
          setIsDialogTyping(false);
        }
      };
      typingTimerId = setTimeout(typeCharacter, TYPING_DELAY_START_MS);
    } else {
      setIsDialogTyping(false);
      if (appState === 'contentReady' || resultsViewActive) {
        if (dialogText !== FULL_DIALOG_TEXT) {
          setDialogText(FULL_DIALOG_TEXT);
        }
      } else {
        if (dialogText !== '') {
          setDialogText('');
        }
      }
    }
    return () => {
      clearTimeout(typingTimerId);
    };
  }, [appState, resultsViewActive]);


  const handleSearchSubmit = async (searchParams) => {
    setCurrentParams(searchParams);
    setIsLoading(true);
    setError(null);
    setSearchResultsData(null);
    setResultsViewActive(true);

    try {
      const { target, algo, mode, max } = searchParams;
      const data = await findRecipes(target, algo, mode, max);
      setSearchResultsData(data);
    } catch (err) {
      setError(err.message || 'Terjadi kesalahan saat mencari resep.');
      setSearchResultsData(null);
    } finally {
      setIsLoading(false);
    }
  };

  let sansClasses = "sans-image";
  if (appState === 'sansAppearing' || appState === 'contentReady' || resultsViewActive) {
    sansClasses += " visible";
  }
  if (resultsViewActive) {
    sansClasses += " results-mode";
  }

  const showInitialElements = appState === 'sansAppearing' || appState === 'contentReady' || resultsViewActive;
  const showSearchForm = appState === 'contentReady' || resultsViewActive;
  const appContainerClasses = `App-container current-state-${appState} ${resultsViewActive ? 'results-view-active' : ''}`;

  return (
    <div className={appContainerClasses}>
      <img
        src={jumboKocokLogo}
        alt="Jumbo Kocok Logo"
        className={`main-logo ${logoPositionClass} ${resultsViewActive ? 'results-mode' : ''}`}
      />

      {appState === 'splash' && (
        <div className="splash-initial-content" onClick={handleInitialClick}>
          <img src={clickToStartImage} alt="Click To Start" className="splash-start-text" />
        </div>
      )}

      <div className={`left-panel ${resultsViewActive ? 'active' : ''}`}>
        {showInitialElements && (
          <>
            <img
              src={sansImage}
              alt="Sans"
              className={sansClasses}
            />
            <div className={`dialog-box ${(appState === 'sansAppearing' || appState === 'contentReady' || resultsViewActive) ? 'visible' : ''} ${resultsViewActive ? 'results-mode' : ''}`}>
              <p className="dialog-text">
                {dialogText}
                {isDialogTyping && <span className="typing-cursor-char">|</span>}
              </p>
            </div>
          </>
        )}
        {showSearchForm && ( // Pastikan ini di-uncomment dan kelas 'visible' ditambahkan dengan benar
          <div className={`search-form-wrapper ${(appState === 'contentReady' || resultsViewActive) ? 'visible' : ''} ${resultsViewActive ? 'results-mode' : ''}`}>
            <SearchForm onSearchSubmit={handleSearchSubmit} isLoading={isLoading} />
          </div>
        )}
      </div>

      {resultsViewActive && ( // Pastikan ini di-uncomment jika ingin melihat hasil
        <div className="right-panel">
          <SearchResults results={searchResultsData} isLoading={isLoading} error={error} />
        </div>
      )}
    </div>
  );
}

export default App;
