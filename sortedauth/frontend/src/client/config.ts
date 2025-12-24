interface Config {
  googleClientId: string;
  googleRedirectUrl: string;
  googleOauthUrl: string;
  appName: string;
  appUrl: string;
}

export const config: Config = {
  appName: 'SortedAuth',
  appUrl: 'http://localhost:3000/',
  googleClientId: 'fake_client_id',
  googleRedirectUrl: '/hack/callback',
  googleOauthUrl: '/hack/fakeoauth/oauth2/v2/auth',
}