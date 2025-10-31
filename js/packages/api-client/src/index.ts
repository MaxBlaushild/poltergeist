import axios, { AxiosInstance } from 'axios';
import { UserZoneReputation } from '@poltergeist/types';

interface Location {
  latitude: number | null;
  longitude: number | null;
  accuracy: number | null;
}

export class APIClient {
  private client: AxiosInstance;
  private getLocation?: () => Location | null;

  constructor(baseURL: string, getLocation?: () => Location | null) {
    this.getLocation = getLocation;
    this.client = axios.create({
      baseURL,
    });

    this.client.interceptors.request.use(
      (config) => {
        const token = localStorage.getItem('token');
        if (token) {
          config.headers['Authorization'] = `Bearer ${token}`;
        }
        
        // Add location header if location is available
        if (this.getLocation) {
          const location = this.getLocation();
          if (location && location.latitude && location.longitude) {
            const locationHeader = `${location.latitude},${location.longitude},${location.accuracy || 0}`;
            config.headers['X-User-Location'] = locationHeader;
          } else {
          }
        } else {
        }
        
        return config;
      },
      (error) => Promise.reject(error)
    );

    this.client.interceptors.response.use(
      (response) => response,
      (error) => {
        if (error.response?.status === 401 || error.response?.status === 403) {
          // Clear invalid token
          localStorage.removeItem('token');
          
          // Get current path
          const currentPath = window.location.pathname;
          
          // Don't redirect if already on login or home page (prevent loops)
          if (currentPath !== '/login' && currentPath !== '/') {
            // Redirect to login with return URL
            window.location.href = `/login?from=${encodeURIComponent(currentPath)}`;
          }
        }
        return Promise.reject(error);
      }
    );
  }

  async get<T>(url: string, params?: any): Promise<T> {
    const response = await this.client.get<T>(url, { params });
    return response.data;
  }

  async post<T>(url: string, data?: any): Promise<T> {
    const response = await this.client.post<T>(url, data);
    return response.data;
  }

  async put<T>(url: string, data?: any): Promise<T> {
    const response = await this.client.put<T>(url, data);
    return response.data;
  }

  async patch<T>(url: string, data?: any): Promise<T> {
    const response = await this.client.patch<T>(url, data);
    return response.data;
  }
  
  async delete<T>(url: string, data?: any): Promise<T> {
    const response = await this.client.delete<T>(url, { data });
    return response.data;
  }

  async getUserReputations(): Promise<UserZoneReputation[]> {
    return this.get<UserZoneReputation[]>('/sonar/reputations');
  }
}

export default APIClient;
