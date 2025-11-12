import { useState } from 'react'
import { sortedtestClient, CreateTenantRequest, CreateProjectRequest, CreateTaskRequest } from '../proto/multi-tenant';

function Home() {
  const [tenantName, setTenantName] = useState('')
  const [createMsg, setCreateMsg] = useState('')
  const [projectName, setProjectName] = useState('')
  const [projectMsg, setProjectMsg] = useState('')
  const [tenantId, setTenantId] = useState('')
  const [taskName, setTaskName] = useState('')
  const [taskMsg, setTaskMsg] = useState('')
  const [projectIdForTask, setProjectIdForTask] = useState('')

  const client = new sortedtestClient('http://localhost:8080')

  const handleCreateTenant = async () => {
    console.info("hi")
    setCreateMsg('')
    const req = new CreateTenantRequest({ name: tenantName })
    try {
      const resp = await client.CreateTenant(req, null)
      setCreateMsg(resp.message)
      setTenantId(resp.message)
    } catch (err) {
      setCreateMsg('Error: ' + (err.message || err.toString()))
    }
  }

  const handleCreateProject = async () => {
    setProjectMsg('')
    const req = new CreateProjectRequest({ name: projectName })
    try {
      const resp = await client.CreateProject(req, { 'tenant-id': tenantId })
      setProjectMsg(resp.message)
    } catch (err) {
      setProjectMsg('Error: ' + (err.message || err.toString()))
    }
  }

  const handleCreateTask = async () => {
    setTaskMsg('')
    const req = new CreateTaskRequest({ project_id: projectIdForTask, name: taskName })
    try {
      const resp = await client.CreateTask(req, { 'tenant-id': tenantId })
      setTaskMsg(resp.message)
    } catch (err) {
      setTaskMsg('Error: ' + (err.message || err.toString()))
    }
  }

  return (
    <div>
      <h2>Create Tenant</h2>
      <input type="text" placeholder="Tenant name" value={tenantName} onChange={e => setTenantName(e.target.value)} />
      <button onClick={handleCreateTenant} style={{ marginLeft: '1em' }}>Create Tenant</button>
      {createMsg && <div style={{ marginTop: '1em', color: 'green' }}>{createMsg}</div>}

      <h2 style={{marginTop:'2em'}}>Create Project</h2>
      <input type="text" placeholder="Tenant ID" value={tenantId} onChange={e => setTenantId(e.target.value)} style={{ marginRight: '1em' }} />
      <input type="text" placeholder="Project name" value={projectName} onChange={e => setProjectName(e.target.value)} />
      <button onClick={handleCreateProject} style={{ marginLeft: '1em' }}>Create Project</button>
      {projectMsg && <div style={{ marginTop: '1em', color: 'blue' }}>{projectMsg}</div>}

      <h2 style={{marginTop:'2em'}}>Create Task</h2>
      <input type="text" placeholder="Project ID" value={projectIdForTask} onChange={e => setProjectIdForTask(e.target.value)} style={{ marginRight: '1em' }} />
      <input type="text" placeholder="Task name" value={taskName} onChange={e => setTaskName(e.target.value)} />
      <button onClick={handleCreateTask} style={{ marginLeft: '1em' }}>Create Task</button>
      {taskMsg && <div style={{ marginTop: '1em', color: 'purple' }}>{taskMsg}</div>}
    </div>
  )
}

export default Home
