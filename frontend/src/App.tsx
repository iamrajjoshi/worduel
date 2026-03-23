import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { GameProvider } from './context/GameContext'
import { LandingPage } from './pages/LandingPage'
import { GamePage } from './pages/GamePage'
import { ResultsPage } from './pages/ResultsPage'

export function App() {
  return (
    <BrowserRouter>
      <GameProvider>
        <Routes>
          <Route path="/" element={<LandingPage />} />
          <Route path="/room/:roomCode" element={<GamePage />} />
          <Route path="/results/:roomCode" element={<ResultsPage />} />
        </Routes>
      </GameProvider>
    </BrowserRouter>
  )
}
