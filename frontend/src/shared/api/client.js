import axios from 'axios';

// Generate lightweight client-side UUID v4 for request correlation
const generateUUID = () => {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0;
    const v = c === 'x' ? r : (r & 0x3) | 0x8;
    return v.toString(16);
  });
};

const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api/v1',
  timeout: 15000,
  headers: {
    'Content-Type': 'application/json',
    Accept: 'application/json',
  },
  withCredentials: true,
});

// Request interceptor: attach correlation ID
apiClient.interceptors.request.use(
  (config) => {
    if (!config.headers['X-Request-ID']) {
      config.headers['X-Request-ID'] = generateUUID();
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// Response interceptor: unwrap backend envelope { data, meta, error }
apiClient.interceptors.response.use(
  (response) => {
    // If the backend envelope is present, return data directly while attaching meta
    if (response.data && 'data' in response.data) {
      return response.data.data;
    }
    return response.data;
  },
  (error) => {
    // Extract domain error payload if present
    if (error.response?.data?.error) {
      const { code, message, details } = error.response.data.error;
      const customErr = new Error(message || 'Request failed');
      customErr.code = code;
      customErr.details = details;
      customErr.status = error.response.status;
      return Promise.reject(customErr);
    }
    return Promise.reject(error);
  }
);

export default apiClient;
