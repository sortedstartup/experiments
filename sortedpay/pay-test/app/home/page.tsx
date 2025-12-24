'use client';
import { useEffect } from "react";



export default function HomePage() {
    useEffect(() => {
        console.log('Home page');

        console.log(localStorage.getItem('sortedchat.jwt'));
    }, []);

    const getUsers = async () => {
      const token = localStorage.getItem('sortedchat.jwt');
      const response = await fetch('/api/home', {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });
      const data = await response.json();
      console.log('data from backend', data);
    }


  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100">
      <div className="text-center">
        <h1 className="text-6xl font-bold text-gray-800 mb-4">Home</h1>
        <p className="text-lg text-gray-600">Welcome! You have successfully logged in.</p>
        <div className="mt-8">

          <button onClick={() => getUsers()} className="inline-block px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors duration-200">
            Get Users
          </button>
          {/* <a 
            href="/login" 
            className="inline-block px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors duration-200"
          >
            Back to Login
          </a> */}
        </div>
      </div>
    </div>
  );
}