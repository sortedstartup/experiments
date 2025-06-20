import axios from 'axios';
import { loadStripe } from '@stripe/stripe-js';

const stripePromise = loadStripe('pk_test_51RbKPJImVq3vV0Ha7QXeRHSb8Yk5bz8suPnrvA3enB0bE9e2NKfofngxB07MQ1DswX5wfH5MPBmkJEk0lZEmJaqu008CSUhtQ8'); 

export default function Home() {
  const createSession = async (type) => {
    console.log("Create session called for type:", type);
  try {
    const endpoint = type === 'subscription'
      ? '/api/payment/subscription-session'
      : '/api/payment/checkout-session';

    const backendUrl = import.meta.env.VITE_BACKEND_URL;
    console.log("Using backend:", import.meta.env.VITE_BACKEND_URL);
    if (!backendUrl) {
      throw new Error("VITE_BACKEND_URL is not defined. Check your .env file.");
    }

    const res = await axios.post(`${backendUrl}${endpoint}`);

    const stripe = await stripePromise;
    await stripe.redirectToCheckout({ sessionId: res.data.sessionId });
  } catch (err) {
    console.error('Error creating session:', err);
    alert('Something went wrong');
  }
};


  return (
    <div className="p-6 space-y-4 text-center">
      <h1 className="text-3xl font-bold">Stripe Payment Demo</h1>
      <button type="button" onClick={() => createSession('one-time')} className="bg-blue-600 text-white px-6 py-2 rounded">
        Buy Now (One-Time)
      </button>
      <button type="button" onClick={() => createSession('subscription')} className="bg-green-600 text-white px-6 py-2 rounded ml-4">
        Subscribe Monthly
      </button>
    </div>
  );
}
