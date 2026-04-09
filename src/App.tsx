import { useCapture } from './store/useCapture'
import PageTabs from './components/PageTabs'
import StartCapture from './pages/StartCapture'
import CaptureDisplay from './pages/CaptureDisplay'
import KeyColors from './pages/KeyColors'
import './App.css'


function App() {
  const { activePage } = useCapture();

  const renderPage = () => {
    switch (activePage) {
      case 'START':
        return <StartCapture />;
      case 'DISPLAY':
        return <CaptureDisplay />;
      case 'COLORS':
        return <KeyColors />;
      default:
        return <StartCapture />;
    }
  };

  return (
    <div className="app-container">
      <header className="app-header">
        <h1>Mini Motorways Solver</h1>
        <p className="subtitle">Optimize your city layout with agentic AI</p>
      </header>

      <PageTabs />

      <main>
        {renderPage()}
      </main>


      <div className="ticks"></div>
      <section id="spacer"></section>
    </div>
  )
}

export default App
