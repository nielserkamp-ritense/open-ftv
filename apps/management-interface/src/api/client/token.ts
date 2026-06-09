let accessToken: string | undefined;
export const setAccessToken = (t?: string): void => { accessToken = t; };
export const getAccessToken = (): string | undefined => accessToken;
