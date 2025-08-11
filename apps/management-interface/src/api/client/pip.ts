import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios';

// Create a PAP-specific API configuration
const papApiConfig: AxiosRequestConfig = {
  baseURL: (import.meta.env.VITE_PIP_BASE_URL as string),
  headers: {
    'Content-Type': 'application/json',
  },
};

// Create the PAP API instance
const pipApiClient: AxiosInstance = axios.create(papApiConfig);

// Request interceptor (auth, etc.)
pipApiClient.interceptors.request.use(
  (config) => {
    // Add authentication headers here if needed
    return config;
  }
);

// Response interceptor for handling errors
pipApiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error instanceof Error) {
      return Promise.reject(error);
    }
    return Promise.reject(new Error(String(error)));
  }
);

// Generic request method with types for PAP
export const pipRequest = async <T>(config: AxiosRequestConfig): Promise<T> => {
  const response: AxiosResponse<T> = await pipApiClient(config);
  return response.data;
};

export default pipApiClient;
