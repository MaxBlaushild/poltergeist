var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
import axios from 'axios';
export class APIClient {
    constructor(baseURL, getLocation) {
        this.getLocation = getLocation;
        this.client = axios.create({
            baseURL,
        });
        this.client.interceptors.request.use((config) => {
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
                }
                else {
                }
            }
            else {
            }
            return config;
        }, (error) => Promise.reject(error));
        this.client.interceptors.response.use((response) => response, (error) => {
            var _a, _b;
            if (((_a = error.response) === null || _a === void 0 ? void 0 : _a.status) === 401 || ((_b = error.response) === null || _b === void 0 ? void 0 : _b.status) === 403) {
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
        });
    }
    get(url, params) {
        return __awaiter(this, void 0, void 0, function* () {
            const response = yield this.client.get(url, { params });
            return response.data;
        });
    }
    post(url, data) {
        return __awaiter(this, void 0, void 0, function* () {
            const response = yield this.client.post(url, data);
            return response.data;
        });
    }
    put(url, data) {
        return __awaiter(this, void 0, void 0, function* () {
            const response = yield this.client.put(url, data);
            return response.data;
        });
    }
    patch(url, data) {
        return __awaiter(this, void 0, void 0, function* () {
            const response = yield this.client.patch(url, data);
            return response.data;
        });
    }
    delete(url, data) {
        return __awaiter(this, void 0, void 0, function* () {
            const response = yield this.client.delete(url, { data });
            return response.data;
        });
    }
    getUserReputations() {
        return __awaiter(this, void 0, void 0, function* () {
            return this.get('/sonar/reputations');
        });
    }
}
export default APIClient;
