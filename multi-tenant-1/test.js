import grpc from "k6/net/grpc";
import { check } from "k6";

export const options = {
  vus: 1,
  duration: "10s",
};

const client = new grpc.Client();
client.load(["proto"], "multi-tenant.proto");

export default () => {
  client.connect("localhost:8000", { plaintext: true });

  // 1. CreateTenant (once)
  const tenantName = "Tenant-" + __ITER;
  const tenantRes = client.invoke("sortedtest.sortedtest/CreateTenant", {
    name: tenantName,
  });
  check(tenantRes, {
    "CreateTenant status is OK": (r) => r && r.status === grpc.StatusOK,
  });

  const params = {
    metadata: { "tenant-id": tenantRes.message && tenantRes.message.message },
    tags: { k6test: "yes" },
  };

  let lastProjectId;
  for (let i = 0; i < 4; i++) {
    const projectRes = client.invoke(
      "sortedtest.sortedtest/CreateProject",
      { name: `Project-${i}` },
      params
    );
    check(projectRes, {
      [`CreateProject #${i} status is OK`]: (r) =>
        r && r.status === grpc.StatusOK,
    });
    if (projectRes && projectRes.message && projectRes.message.message) {
      lastProjectId = projectRes.message.message;
    }
  }

  for (let j = 0; j < 2; j++) {
    const taskRes = client.invoke(
      "sortedtest.sortedtest/CreateTask",
      {
        project_id: lastProjectId,
        name: `Task-${j}`,
      },
      params
    );
    check(taskRes, {
      [`CreateTask #${j} status is OK`]: (r) => r && r.status === grpc.StatusOK,
    });
  }

  const getProjectsRes = client.invoke(
    "sortedtest.sortedtest/GetProjects",
    {},
    params
  );
  check(getProjectsRes, {
    "GetProjects status is OK": (r) => r && r.status === grpc.StatusOK,
  });

  const getTasksRes = client.invoke(
    "sortedtest.sortedtest/GetTasks",
    { project_id: lastProjectId },
    params
  );
  check(getTasksRes, {
    "GetTasks status is OK": (r) => r && r.status === grpc.StatusOK,
  });

  client.close();
};
