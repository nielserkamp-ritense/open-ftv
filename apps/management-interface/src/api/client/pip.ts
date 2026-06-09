import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios';
import { PIP_BASE_URL } from '@/config/env';
import { getAccessToken } from './token';

// Create a PAP-specific API configuration
const papApiConfig: AxiosRequestConfig = {
  baseURL: PIP_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
};

// Create the PAP API instance
const pipApiClient: AxiosInstance = axios.create(papApiConfig);

// Request interceptor (auth, etc.)
pipApiClient.interceptors.request.use((config) => {
  const token = getAccessToken();
  if (token) { config.headers.Authorization = `Bearer ${token}`; }
  return config;
});

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
