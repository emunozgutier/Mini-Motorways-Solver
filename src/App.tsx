import { useCapture } from './store/useCapture'
import PageTabs from './components/PageTabs'
import StartCapture from './pages/StartCapture'
import CaptureDisplay from './pages/CaptureDisplay'
import './App.css'

function App() {
  const { activePage } = useCapture();

  return (
    <div className="app-container">
      <header className="app-header">
        <h1>Mini Motorways Solver</h1>
        <p className="subtitle">Optimize your city layout with agentic AI</p>
      </header>

      <PageTabs />

      <main>
        {activePage === 'START' ? (
          <StartCapture />
        ) : (
          <CaptureDisplay />
        )}
      </main>

      <div className="ticks"></div>
      <section id="spacer"></section>
    </div>
  )
}

export default App
