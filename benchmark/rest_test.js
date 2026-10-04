import http from 'k6/http';
import { check } from 'k6';

export default function () {
  const res = http.get('http://localhost:8081/api/v1/stock?product_id=PROD-001');
  check(res, {
    'status is 200': (r) => r.status === 200,
  });
}