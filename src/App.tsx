import { useState } from 'react'
import CaptureFlow from './components/CaptureFlow'
import './App.css'

function App() {
  const [capture, setCapture] = useState<{
    stream: MediaStream;
    cropArea: { x: number, y: number, width: number, height: number } | null;
  } | null>(null);

  return (
    <div className="app-container">
      <header className="app-header">
        <h1>Mini Motorways Solver</h1>
        <p className="subtitle">Optimize your city layout with agentic AI</p>
      </header>

      <main>
        {!capture ? (
          <CaptureFlow 
            onCaptureComplete={(stream, cropArea) => setCapture({ stream, cropArea })} 
          />
        ) : (
          <div className="solver-active">
            <div className="status-badge">Solver Active</div>
            {/* Solver UI components will go here */}
            <button className="btn-secondary" onClick={() => setCapture(null)}>
              Stop Capture
            </button>
          </div>
        )}
      </main>

      <div className="ticks"></div>
      <section id="spacer"></section>
    </div>
  )
}

export default App
