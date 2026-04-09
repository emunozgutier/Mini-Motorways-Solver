import React from 'react';
import { useCapture } from '../store/useCapture';

const PageTabs: React.FC = () => {
  const { activePage, setActivePage, step } = useCapture();
  
  const isDisplayDisabled = step !== 'READY';

  return (
    <nav className="tab-nav">
      <button 
        className={`tab-item ${activePage === 'START' ? 'active' : ''}`}
        onClick={() => setActivePage('START')}
      >
        <span className="icon">⚙️</span>
        Setup
      </button>
      <button 
        className={`tab-item ${activePage === 'COLORS' ? 'active' : ''} ${isDisplayDisabled ? 'disabled' : ''}`}
        onClick={() => !isDisplayDisabled && setActivePage('COLORS')}
        title={isDisplayDisabled ? 'Complete the setup first' : ''}
      >
        <span className="icon">🎨</span>
        Colors
      </button>
      <button 
        className={`tab-item ${activePage === 'DISPLAY' ? 'active' : ''} ${isDisplayDisabled ? 'disabled' : ''}`}
        onClick={() => !isDisplayDisabled && setActivePage('DISPLAY')}
        title={isDisplayDisabled ? 'Complete the setup first' : ''}
      >
        <span className="icon">📺</span>
        Display
      </button>

    </nav>
  );
};

export default PageTabs;
