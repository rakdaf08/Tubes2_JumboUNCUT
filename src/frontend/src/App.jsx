// src/frontend/src/App.jsx
import React, { useState, useEffect } from 'react';
import './App.css';
import SearchForm from './components/SearchForm';
import SearchResults from './components/SearchResults';
import { findRecipes } from './api/searchService';

import jumboKocokLogo from './assets/Jumbo_Kocok.png';
import clickToStartImage from './assets/ClickToStart.png';
import sansImage from './assets/sans.jpg';

const FULL_DIALOG_TEXT = "* Please enter the recipe you are looking for";
const TYPING_DELAY_START_MS = 450;
const TYPING_SPEED_MS = 70;

function App() {
  const [appState, setAppState] = useState('splash'); 
  const [logoPositionClass, setLogoPositionClass] = useState('center'); 
  const [resultsViewActive, setResultsViewActive] = useState(false);

  const [searchResultsData, setSearchResultsData] = useState(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);
  const [currentParams, setCurrentParams] = useState(null);

  const [dialogText, setDialogText] = useState('');
  const [isDialogTyping, setIsDialogTyping] = useState(false);
  const [isInitialTypingComplete, setIsInitialTypingComplete] = useState(false); 

  const handleInitialClick = () => {
    if (appState === 'splash') {
      setLogoPositionClass('topLeft');
      setAppState('logoMoving');
    }
  };

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
      const searchFormTransitionDelay = 1200;

      const timeUntilContentReady = Math.max(
        dialogBoxAnimationDuration,
        sansAppearanceTime,
        totalTypingTimeMs,
        searchFormTransitionDelay
      ) + 200; 

      timer = setTimeout(() => {
        setAppState('contentReady');
      }, timeUntilContentReady);
    }
    return () => clearTimeout(timer);
  }, [appState]);

  useEffect(() => {
    let typingTimerId;

    if (appState === 'sansAppearing' && !isInitialTypingComplete) {
      setDialogText(''); 
      setIsDialogTyping(true);
      let currentCharIndex = 0; 

      const typeCharacter = () => {
        if (currentCharIndex < FULL_DIALOG_TEXT.length) {
          setDialogText(prev => prev + FULL_DIALOG_TEXT.charAt(currentCharIndex));
          currentCharIndex++;
          typingTimerId = setTimeout(typeCharacter, TYPING_SPEED_MS);
        } else {
          setIsDialogTyping(false);
          setIsInitialTypingComplete(true); 
        }
      };
      typingTimerId = setTimeout(typeCharacter, TYPING_DELAY_START_MS);

    } else if (appState === 'contentReady') {
      if (!isInitialTypingComplete) {
        setDialogText(FULL_DIALOG_TEXT); 
        setIsDialogTyping(false);
        setIsInitialTypingComplete(true); 
      } else {
        setIsDialogTyping(false);
      }
    } else if (appState === 'splash' || appState === 'logoMoving') {
      setDialogText('');
      setIsDialogTyping(false);
      setIsInitialTypingComplete(false); 
    } else if (resultsViewActive) {
      setDialogText(FULL_DIALOG_TEXT);
      setIsDialogTyping(false);
      if (!isInitialTypingComplete) {
        setIsInitialTypingComplete(true);
      }
    }

    return () => {
      clearTimeout(typingTimerId); 
    };
  }, [appState, resultsViewActive, isInitialTypingComplete]);


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

  const showSplashContent = appState === 'splash';
  const showMainInteractiveElements = appState === 'sansAppearing' || appState === 'contentReady' || resultsViewActive;
  const showSearchForm = appState === 'contentReady' || resultsViewActive;

  const appContainerClasses = `App-container current-state-${appState} ${resultsViewActive ? 'results-view-active' : ''}`;

  return (
    <div className={appContainerClasses}>
      <img
        src={jumboKocokLogo}
        alt="Jumbo Kocok Logo"
        className={`main-logo ${logoPositionClass} ${resultsViewActive ? 'results-mode' : ''}`}
      />

      {showSplashContent && (
        <div className="splash-initial-content" onClick={handleInitialClick}>
          <img src={clickToStartImage} alt="Click To Start" className="splash-start-text" />
        </div>
      )}

      <div className={`left-panel ${resultsViewActive ? 'active' : ''}`}>
        {showMainInteractiveElements && (
          <>
            <img
              src={sansImage}
              alt="Sans"
              className={sansClasses}
            />
            <div className={`dialog-and-form-container ${resultsViewActive ? 'results-mode' : ''} ${(appState === 'sansAppearing' || appState === 'contentReady' || resultsViewActive) ? 'visible' : ''}`}>
              <div className={`dialog-box ${resultsViewActive ? 'results-mode' : ''}`}>
                <p className="dialog-text">
                  {dialogText}
                  {isDialogTyping && <span className="typing-cursor-char">|</span>}
                </p>
              </div>

              {showSearchForm && (
                <div className={`search-form-wrapper ${resultsViewActive ? 'results-mode' : ''}`}>
                  <SearchForm onSearchSubmit={handleSearchSubmit} isLoading={isLoading} />
                </div>
              )}
            </div>
          </>
        )}
      </div>

      {resultsViewActive && (
        <div className="right-panel">
          <SearchResults results={searchResultsData} isLoading={isLoading} error={error} />
        </div>
      )}
    </div>
  );
}

export default App;
