import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import Home from './Home.jsx'
import Projects from './Projects.jsx'
import { BrowserRouter, Routes, Route, Link, createRoutesFromChildren, matchRoutes, useLocation, useNavigationType } from 'react-router-dom';
import { ConsoleInstrumentation, createReactRouterV6Options, getWebInstrumentations, initializeFaro, ReactIntegration, LogLevel } from '@grafana/faro-react';
import { TracingInstrumentation } from '@grafana/faro-web-tracing';

initializeFaro({
  url: 'https://faro-collector-prod-ap-south-1.grafana.net/collect/4eef15fc7fecfdfce6b86796ee3825a7',
  app: {
    name: 'Full Stack Test',
    version: '1.0.0',
    environment: 'production'
  },
  instrumentations: [
    ...getWebInstrumentations(),
    new TracingInstrumentation({
      instrumentationOptions: {
        propagateTraceHeaderCorsUrls: [new RegExp('http://localhost:8080/.*')],
      },
    }),
    new ReactIntegration({
      router: createReactRouterV6Options({
        createRoutesFromChildren,
        matchRoutes,
        Routes,
        useLocation,
        useNavigationType,
      }),
    }),
  ],
});

createRoot(document.getElementById('root')).render(
  <StrictMode>
    <BrowserRouter>
      <nav style={{ margin: '1em' }}>
        <Link to="/" style={{ marginRight: '1em' }}>Home</Link>
        <Link to="/projects">Projects</Link>
      </nav>
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/projects" element={<Projects />} />
      </Routes>
    </BrowserRouter>
  </StrictMode>,
)
