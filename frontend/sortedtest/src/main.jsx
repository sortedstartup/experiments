import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import './index.css'
import App from './App.jsx'
import { createRoutesFromChildren, matchRoutes, Routes, useLocation, useNavigationType } from 'react-router-dom';
import { ConsoleInstrumentation, createReactRouterV6Options, getWebInstrumentations, initializeFaro, ReactIntegration, ReactRouterVersion, LogLevel } from '@grafana/faro-react';
import { TracingInstrumentation } from '@grafana/faro-web-tracing';

initializeFaro({
  url: 'https://faro-collector-prod-ap-south-1.grafana.net/collect/2423025ccb0d551e822adc5396715638',
  app: {
    name: 'Sanskar-Sorted-Test',
    version: '1.0.0',
    environment: 'production'
  },
  
  instrumentations: [


    ...getWebInstrumentations(),
    
    new TracingInstrumentation(),


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
      <App />
    </BrowserRouter>
  </StrictMode>,
)
