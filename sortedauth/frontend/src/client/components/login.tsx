import { config } from '../config.ts';

export default function Login() {

  const handleGoogleLogin = () => {
    try {
      // Get configuration from Vite environment variables
      const clientId = config.googleClientId;
      const redirectUri = config.googleRedirectUrl;
      if (!clientId) {
        console.error('Client ID is not configured.');
        return;
      }
      // Build Google OAuth URL directly (same parameters as backend /login)
      const googleOAuthParams = new URLSearchParams({
        client_id: clientId,
        redirect_uri: redirectUri,
        response_type: 'code',
        scope: 'openid email profile',
        access_type: 'offline',
        state: 'state', // In production, this should be a random value for CSRF protection
      } as Record<string, string>);
      const googleOAuthURL = `${config.googleOauthUrl}?${googleOAuthParams.toString()}`;
      // Redirect to Google OAuth
      window.location.href = googleOAuthURL;
    } catch (err) {
      console.error('Login error:', err);
    }
  }
  return (
    <>
      <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-slate-50 to-slate-100 py-12 px-4 sm:px-6 lg:px-8">
        <div className="max-w-md w-full space-y-8">
          {/* Logo/Brand Section */}
          <div className="text-center">
            <div className="mx-auto h-12 w-12 bg-gradient-to-br from-blue-600 to-purple-600 rounded-xl flex items-center justify-center mb-6">
              <svg className="h-6 w-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
              </svg>
            </div>
            <h2 className="text-3xl font-bold tracking-tight text-gray-900">
              Welcome to {config.appName}
            </h2>
            <p className="mt-2 text-sm text-gray-600">
              Sign in with your Google account to continue
            </p>
          </div>

          <div className="p-8 shadow-xl border-0 bg-white/80 backdrop-blur-sm rounded-lg">

            <div className="space-y-6">
              <button
                onClick={handleGoogleLogin}
                className="w-full h-12 bg-white hover:bg-gray-50 text-gray-700 border border-gray-300 shadow-sm transition-all duration-200 hover:shadow-md rounded-lg disabled:opacity-50 disabled:cursor-not-allowed"
              >

                <div className="flex items-center justify-center space-x-3">
                  <svg className="w-5 h-5" viewBox="0 0 24 24">
                    <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" />
                    <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" />
                    <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" />
                    <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" />
                  </svg>
                  <span className="font-medium">Continue with Google</span>
                </div>

              </button>

              <div className="relative">
                <div className="absolute inset-0 flex items-center">
                  <div className="w-full border-t border-gray-200"></div>
                </div>
                <div className="relative flex justify-center text-sm">
                  <span className="px-2 bg-white text-gray-500">Secure OAuth Authentication</span>
                </div>
              </div>

              <div className="text-xs text-gray-500 text-center space-y-1">
                <p>By signing in, you agree to our terms of service.</p>
                <p>Your data is protected with enterprise-grade security.</p>
              </div>
            </div>
          </div>

          {/* Footer */}
          <div className="text-center text-xs text-gray-400">
            <p>© 2024 {config.appName}. All rights reserved.</p>
          </div>
        </div>
      </div>
    </>

  )
}
