import { BrowserRouter, Route, Routes } from 'react-router-dom'
import Header from './components/Header'
import { AuthProvider } from './lib/auth'
import BoardPage from './pages/BoardPage'
import HomePage from './pages/HomePage'
import LoginPage from './pages/LoginPage'
import RegisterPage from './pages/RegisterPage'
import ThreadPage from './pages/ThreadPage'
import UserPage from './pages/UserPage'

function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <Header />
        <main className="main">
          <Routes>
            <Route path="/" element={<HomePage />} />
            <Route path="/b/:board" element={<BoardPage />} />
            <Route path="/t/:id" element={<ThreadPage />} />
            <Route path="/users/:username" element={<UserPage />} />
            <Route path="/login" element={<LoginPage />} />
            <Route path="/register" element={<RegisterPage />} />
          </Routes>
        </main>
      </BrowserRouter>
    </AuthProvider>
  )
}

export default App
