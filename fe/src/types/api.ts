export type TApiErrorBody = {
  error: {
    code: string;
    message: string;
    params?: Record<string, unknown>;
    fields?: Record<string, string>;
  };
};

export type TSessionUser = {
  id: string;
  email: string;
  display_name: string;
  avatar_url: string;
  ui_locale: "vi" | "en";
};

export type TSession = {
  access_token: string;
  access_token_expires_at: string;
  refresh_token: string;
  refresh_token_expires_at: string;
  user: TSessionUser;
};
