import "server-only";

export const serverEnv = {
  apiUrl: (process.env.API_URL ?? "http://localhost:8080").replace(/\/$/, ""),
  googleClientId: process.env.OAUTH_CLIENT_ID ?? "",
  googleRedirectUrl:
    process.env.OAUTH_REDIRECT_URL ?? "http://localhost:3000/api/auth/google/callback",
};
