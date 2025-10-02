
import grpc from 'k6/net/grpc';
import { check, sleep } from 'k6';

export const options = {
  vus: 2,
  duration: '10s',
};

const client = new grpc.Client();
client.load(['./proto'], 'otel.proto');


export default () => {
  client.connect('localhost:8000', {
    plaintext: true, 
  });

   const data = {
    chatId: 'test-chat',
    message: 'k6'
  };

  const res = client.invoke('sortedtest.sortedtest/test', data);

  check(res, {
    'status is OK': (r) => r && r.status === grpc.StatusOK,
  });

  client.close();
  // sleep(1);
};