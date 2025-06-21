import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import Home from './pages/Home';
import Success from './pages/Success';
import Cancel from './pages/Cancel';
import PaidPage from './pages/PaidPage';

function App() {
  return (
    <Router>
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/success" element={<Success />} />
        <Route path="/cancel" element={<Cancel />} />
        <Route path="/paid" element={<PaidPage />} />
      </Routes>
    </Router>
  );
}

export default App;
