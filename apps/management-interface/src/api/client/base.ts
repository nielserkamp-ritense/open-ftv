import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios';

// Create a base API configuration
const apiConfig: AxiosRequestConfig = {
    baseURL: (import.meta.env.VITE_PAP_BASE_URL as string) || '/api',
    headers: {
        'Content-Type': 'application/json',
    },
};

// Create the API instance
const apiClient: AxiosInstance = axios.create(apiConfig);

// Request interceptor for adding auth tokens, etc.
apiClient.interceptors.request.use(
    (config) => {
        // You can add authentication headers here
        // const token = getToken();
        // if (token) {
        //   config.headers.Authorization = `Bearer ${token}`;
        // }
        return config;
    }
);

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