interface Config {
    googleClientId: string;
    googleRedirectUrl: string;
    googleOauthUrl: string;
    appName: string;
    appUrl: string;
    jwtIssuer: string;
}

export const config: Config = {
    appName: 'NextApp',
    appUrl: 'http://localhost:3000',
    googleClientId: '410294925787-nvmuoh607khojahfm5eqtrcu4o0jp87a.apps.googleusercontent.com',
    googleRedirectUrl: 'http://localhost:3000/oauth/callback/google',
    googleOauthUrl: 'https://accounts.google.com/o/oauth2/v2/auth',
    jwtIssuer: 'sanskar',
}