import { useState } from 'react'
import reactLogo from './assets/react.svg'
import viteLogo from '/vite.svg'
import './App.css'
// Import gRPC client and messages
import { sortedtestClient, CreateTenantRequest, CreateProjectRequest, CreateTaskRequest } from '../proto/multi-tenant';

function App() {
  const [count, setCount] = useState(0)
  const [tenantName, setTenantName] = useState('')
  const [createMsg, setCreateMsg] = useState('')
  const [projectName, setProjectName] = useState('')
  const [projectMsg, setProjectMsg] = useState('')
  const [tenantId, setTenantId] = useState('14a12649-4bf6-48c2-b916-ecea32a0c6c7') // new state for tenant id
  const [taskName, setTaskName] = useState('')
  const [taskMsg, setTaskMsg] = useState('')

  // Set your backend gRPC-web proxy address here
  const client = new sortedtestClient('http://localhost:8080')

  const handleCreateTenant = async () => {
    setCreateMsg('')
    const req = new CreateTenantRequest({ name: tenantName })
    try {
      const resp = await client.CreateTenant(req, null)
      setCreateMsg(resp.message)
      setTenantId(resp.message) // set tenantId to the new tenant's id
    } catch (err) {
      setCreateMsg('Error: ' + (err.message || err.toString()))
    }
  }

  const handleCreateProject = async () => {
    setProjectMsg('')
    const req = new CreateProjectRequest({ name: projectName })
    try {
      const resp = await client.CreateProject(req, { 'x-tenant-id': tenantId })
      setProjectMsg(resp.message)
    } catch (err) {
      setProjectMsg('Error: ' + (err.message || err.toString()))
    }
  }

  const handleCreateTask = async () => {
    setTaskMsg('')
    const req = new CreateTaskRequest({ project_id: 'ede2c5a8-a8cc-4eca-8804-c45a1be28db4', name: taskName })
    try {
      const resp = await client.CreateTask(req, { 'x-tenant-id': '14a12649-4bf6-48c2-b916-ecea32a0c6c7' })
      setTaskMsg(resp.message)
    } catch (err) {
      setTaskMsg('Error: ' + (err.message || err.toString()))
    }
  }

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
      <div style={{ margin: '2em 0' }}>
        <h2>Create Tenant</h2>
        <input
          type="text"
          placeholder="Tenant name"
          value={tenantName}
          onChange={e => setTenantName(e.target.value)}
        />
        <button onClick={handleCreateTenant} style={{ marginLeft: '1em' }}>
          Create Tenant
        </button>
        {createMsg && <div style={{ marginTop: '1em', color: 'green' }}>{createMsg}</div>}
      </div>
      <div style={{ margin: '2em 0' }}>
        <h2>Create Project</h2>
        <input
          type="text"
          placeholder="Tenant ID"
          value={tenantId}
          onChange={e => setTenantId(e.target.value)}
          style={{ marginRight: '1em' }}
        />
        <input
          type="text"
          placeholder="Project name"
          value={projectName}
          onChange={e => setProjectName(e.target.value)}
        />
        <button onClick={handleCreateProject} style={{ marginLeft: '1em' }}>
          Create Project
        </button>
        {projectMsg && <div style={{ marginTop: '1em', color: 'blue' }}>{projectMsg}</div>}
      </div>
      <div style={{ margin: '2em 0' }}>
        <h2>Create Task (Tenant & Project hardcoded)</h2>
        <input
          type="text"
          placeholder="Task name"
          value={taskName}
          onChange={e => setTaskName(e.target.value)}
        />
        <button onClick={handleCreateTask} style={{ marginLeft: '1em' }}>
          Create Task
        </button>
        {taskMsg && <div style={{ marginTop: '1em', color: 'purple' }}>{taskMsg}</div>}
      </div>
    </>
  )
}

export default App
