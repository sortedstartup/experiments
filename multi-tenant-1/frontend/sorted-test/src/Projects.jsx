import { useEffect, useState } from 'react'
import { sortedtestClient, GetProjectsRequest, GetTasksRequest } from '../proto/multi-tenant';

function Projects() {
  const [projects, setProjects] = useState([])
  const [_, setSelectedProjectId] = useState('')
  const [tasks, setTasks] = useState([])
  const [tenantId, setTenantId] = useState('')
  const [error, setError] = useState('')

  const client = new sortedtestClient('http://localhost:8080')

  const fetchProjects = async () => {
    setError('')
    try {
      const req = new GetProjectsRequest()
      const resp = await client.GetProjects(req, { 'tenant-id': tenantId })
      setProjects(resp.projects)
    } catch (err) {
      setError('Error fetching projects: ' + (err.message || err.toString()))
    }
  }

  const fetchTasks = async (projectId) => {
    setError('')
    try {
      const req = new GetTasksRequest({ project_id: projectId })
      const resp = await client.GetTasks(req, { 'tenant-id': tenantId })
      setTasks(resp.tasks)
    } catch (err) {
      setError('Error fetching tasks: ' + (err.message || err.toString()))
    }
  }

  useEffect(() => {
    if (tenantId) fetchProjects()
  }, [tenantId])

  return (
    <div>
      <h2>Projects</h2>
      <input type="text" placeholder="Tenant ID" value={tenantId} onChange={e => setTenantId(e.target.value)} />
      <button onClick={fetchProjects} style={{ marginLeft: '1em' }}>Fetch Projects</button>
      {error && <div style={{ color: 'red', marginTop: '1em' }}>{error}</div>}
      <ul>
        {projects && projects.map(p => (
          <li key={p.id}>
            <button onClick={() => { setSelectedProjectId(p.id); fetchTasks(p.id); }}>
              {p.name} (ID: {p.id})
            </button>
          </li>
        ))}
      </ul>
      <h3>Tasks</h3>
      <ul>
        {tasks && tasks.map(t => (
          <li key={t.id}>{t.name} (ID: {t.id})</li>
        ))}
      </ul>
    </div>
  )
}

export default Projects
