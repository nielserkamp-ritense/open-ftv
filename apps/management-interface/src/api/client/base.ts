import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios';
import { PAP_BASE_URL } from '@/config/env';
import { getAccessToken } from './token';

// Create a base API configuration
const apiConfig: AxiosRequestConfig = {
    baseURL: PAP_BASE_URL,
    headers: {
        'Content-Type': 'application/json',
    },
};

// Create the API instance
const apiClient: AxiosInstance = axios.create(apiConfig);

// Request interceptor for adding auth tokens, etc.
apiClient.interceptors.request.use((config) => {
    const token = getAccessToken();
    if (token) { config.headers.Authorization = `Bearer ${token}`; }
    return config;
});

// Response interceptor for handling errors
apiClient.interceptors.response.use(
    (response) => response,
    (error) => {
        // Handle common errors (401, 403, 500, etc.)
        // You could also redirect to login page on auth errors
        if (error instanceof Error) {
            return Promise.reject(error);
        }
        return Promise.reject(new Error(String(error)));
    }
);

// Generic request method with types
export const request = async <T>(config: AxiosRequestConfig): Promise<T> => {
    const response: AxiosResponse<T> = await apiClient(config);
    return response.data;
};

export default apiClient;