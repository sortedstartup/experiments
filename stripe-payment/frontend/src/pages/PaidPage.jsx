import { useEffect, useState } from 'react';
import axios from 'axios';

export default function PaidPage() {
  const [access, setAccess] = useState(null);
  const email = "swapnil9srivastava@gmail.com"; // Ideally from auth

  useEffect(() => {
    axios.get(`${import.meta.env.VITE_BACKEND_URL}/api/payment/is-paid?email=${email}`)
      .then(res => setAccess(res.data.access))
      .catch(() => setAccess(false));
  }, []);

  if (access === null) return <div>Checking access...</div>;
  if (!access) return <div>Access Denied: Please subscribe to view this page.</div>;

  return <div>Welcome to the Paid Content!</div>;
}
