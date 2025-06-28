import { useState, useEffect } from 'react'
import reactLogo from './assets/react.svg'
import viteLogo from '/vite.svg'
import './App.css'
import { sortedtestClient, testRequest } from '../proto/otel'

function App() {
  const [count, setCount] = useState(0)

  useEffect(() => {
    fetch('/api')
      .then(res => res.json())
      .then(data => {
        console.log('API response:', data)
      })
      .catch(err => {
        console.error('API error:', err)
      })

    // gRPC-web client setup
    const client = new sortedtestClient('http://localhost:8080'); // Change to your gRPC-web proxy address
    const req = new testRequest({ message: 'Hello from frontend', chat_id: 'frontend' });
    client.test(req, null)
      .then(res => {
        console.log(res);
        
        console.log('gRPC Test API response:', res.toObject ? res.toObject() : res)
      })
      .catch(err => {
        console.error('gRPC Test API error:', err)
      })
  }, [])

  return (
    <>
      <div>
        <a href="https://vite.dev" target="_blank">
          <img src={viteLogo} className="logo" alt="Vite logo" />
        </a>
        <a href="https://react.dev" target="_blank">
          <img src={reactLogo} className="logo react" alt="React logo" />
        </a>
      </div>
      <h1>Vite + React</h1>
      <div className="card">
        <button onClick={() => setCount((count) => count + 1)}>
          count is {count}
        </button>
        <p>
          Edit <code>src/App.jsx</code> and save to test HMR
        </p>
      </div>
      <p className="read-the-docs">
        Click on the Vite and React logos to learn more
      </p>
    </>
  )
}

export default App
