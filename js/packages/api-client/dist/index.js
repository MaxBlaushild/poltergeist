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
    constructor(baseURL) {
        this.client = axios.create({
            baseURL,
        });
        this.client.interceptors.request.use((config) => {
            const token = localStorage.getItem('token');
            if (token) {
                config.headers['Authorization'] = `Bearer ${token}`;
            }
            return config;
        }, (error) => Promise.reject(error));
    }
    get(url, params) {
        return __awaiter(this, void 0, void 0, function* () {
            console.log(url);
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
    patch(url, data) {
        return __awaiter(this, void 0, void 0, function* () {
            const response = yield this.client.patch(url, data);
            return response.data;
        });
    }
    delete(url) {
        return __awaiter(this, void 0, void 0, function* () {
            const response = yield this.client.delete(url);
            return response.data;
        });
    }
}
export default APIClient;
