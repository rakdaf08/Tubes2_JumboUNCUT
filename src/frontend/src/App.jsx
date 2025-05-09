// src/frontend/src/App.jsx
import React, { useState, useEffect } from 'react';
import './App.css';
import SearchPage from './pages/SearchPage';

import jumboKocokLogo from './assets/Jumbo_Kocok.png';
import clickToStartImage from './assets/ClickToStart.png';
import sansImage from './assets/sans.jpg';

function App() {
  const [appState, setAppState] = useState('splash');
  const [logoPositionClass, setLogoPositionClass] = useState('center');

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
      }, 1000); // Durasi animasi logo
    } else if (appState === 'sansAppearing') {
      // Setelah Sans dan dialog selesai animasi munculnya (sesuai durasi animasi CSS)
      timer = setTimeout(() => {
        setAppState('contentReady'); // State di mana Sans bobbing, dialog terlihat, DAN form muncul
      }, 900); // Sesuaikan delay ini agar Sans & dialog selesai animasi (misal fadeInSans 0.5s, fadeInDialogSmooth 0.3s delay + 0.6s durasi = 0.9s)
    }
    return () => clearTimeout(timer);
  }, [appState]);

  const showMainElements = appState === 'sansAppearing' || appState === 'contentReady';
  const showSearchPageContent = appState === 'contentReady';

  return (
    <div className={`App-container current-state-${appState}`}>
      <img
        src={jumboKocokLogo}
        alt="Jumbo Kocok Logo"
        className={`main-logo ${logoPositionClass}`}
      />

      {appState === 'splash' && (
        <div className="splash-initial-content" onClick={handleInitialClick}>
          <img src={clickToStartImage} alt="Click To Start" className="splash-start-text" />
        </div>
      )}

      {showMainElements && (
        <>
          <img
            src={sansImage}
            alt="Sans"
            className={`sans-image ${appState === 'contentReady' ? 'bobbing visible' : 'visible'}`}
            // 'visible' class untuk fadeInSans, 'bobbing' ditambahkan saat contentReady
          />
          <div className={`dialog-box ${showMainElements ? 'visible' : ''}`}>
            <p className="dialog-text">
              * Please enter the recipe you are looking for
            </p>
          </div>
        </>
      )}

      {/* SearchPage akan dirender dan diberi kelas 'visible' saat contentReady */}
      <div className={`search-page-wrapper ${showSearchPageContent ? 'visible' : ''}`}>
        {showSearchPageContent && <SearchPage />}
      </div>
    </div>
  );
}

export default App;