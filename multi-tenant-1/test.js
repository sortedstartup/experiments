import grpc from 'k6/net/grpc';
import { check } from 'k6';

export const options = {
  vus: 1,           
  duration: '10s',
};

const client = new grpc.Client();
// Adjust this if your proto is in a different folder or named differently
client.load(['proto'], 'multi-tenant.proto');

export default () => {
  client.connect('localhost:8000', {
    plaintext: true,
  });

  // 1. CreateTenant with a unique name per iteration
  const tenantName = 'Tenant-' + __ITER;
  const tenantRes = client.invoke('sortedtest.sortedtest/CreateTenant', {
    name: tenantName,
  });

  check(tenantRes, {
    'CreateTenant status is OK': (r) => r && r.status === grpc.StatusOK,
  });

   const params = {
    metadata: {
      'tenant-id': tenantRes.message.message,
    },
    tags: { k6test: 'yes' },
  };

  // 2. CreateProject (static project name is fine unless you want to parameterize it too)
  const projectRes = client.invoke('sortedtest.sortedtest/CreateProject', {
    name: 'Project-Alpha',
  },params);

  check(projectRes, {
    'CreateProject status is OK': (r) => r && r.status === grpc.StatusOK,
  });

   const params1 = {
    metadata: {
      'tenant-id': tenantRes.message.message,
    },
    tags: { k6test: 'yes' },
  };

  // 3. CreateTask — passing a placeholder project_id
  // You can improve this by returning a project_id from the backend if supported
  const taskRes = client.invoke('sortedtest.sortedtest/CreateTask', {
    project_id: projectRes.message.message, // Replace this with actual ID if you return one
    name: 'Initial Setup Task',
  },params1);

  check(taskRes, {
    'CreateTask status is OK': (r) => r && r.status === grpc.StatusOK,
  });

  client.close();
};
