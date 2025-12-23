'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
interface JWT {
  sub: string;      // user id
  email: string;
  roles: string[];
  iss: string;      // issuer
  exp: number;      // expiration timestamp
  iat: number;      // issued at timestamp
}

export default function Home() {
  const [isLoading, setIsLoading] = useState(true);
  const router = useRouter();

  useEffect(() => {
    const { google } = window as any;

    if (google) {
      google.accounts.id.initialize({
        client_id: '410294925787-nvmuoh607khojahfm5eqtrcu4o0jp87a.apps.googleusercontent.com',
        callback: async (response: any) => {
          console.log('response', response);

          const responseData = {
            token: response.credential,
          };

          const oneTapResponse = await fetch('http://localhost:8080/google-one-tap-callback', {
            method: 'POST',
            body: JSON.stringify(responseData),
          });
          console.log('oneTapResponse', oneTapResponse);
          if (oneTapResponse.ok) {
            const data = await oneTapResponse.json();
            console.log('data', data);
            localStorage.setItem('sortedchat.jwt', data.jwt);
            router.push('/');
          } else {
            console.error('Failed to call Google One Tap callback');
            router.push('/login');
          }
        },
      });
      google.accounts.id.prompt();
    }
  }, []);

  useEffect(() => {
    // Check if user is authenticated
    const token = localStorage.getItem('sortedchat.jwt');
    
    if (token) {
      // User is authenticated, redirect to home page
      router.push('/home');

      function decodeJWTPayload(token: string): JWT | null {
        try {
          const parts = token.split('.');
          if (parts.length !== 3) {
            return null;
          }
          
          const payload = parts[1];
          const decoded = JSON.parse(atob(payload.replace(/-/g, '+').replace(/_/g, '/')));
          return decoded as JWT;
        } catch (error) {
          console.error('Failed to decode JWT payload:', error);
          return null;
        }
      }

      const payload = decodeJWTPayload(token);
      console.log('payload', payload);
      console.log('payload.sub', payload?.sub);
      console.log('payload.email', payload?.email);
      console.log('payload.roles', payload?.roles);
      console.log('payload.iss', payload?.iss);
      console.log('payload.exp', payload?.exp);
      console.log('payload.iat', payload?.iat);
    } else {
      // User is not authenticated, redirect to login
      router.push('/login');
    }
    
    setIsLoading(false);
  }, [router]);

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-100">
        <div className="text-center">
          <div className="w-8 h-8 border-4 border-blue-600 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
          <p className="text-gray-600">Loading...</p>
        </div>
      </div>
    );
  }

  return null; // This won't render as we redirect immediately
}

