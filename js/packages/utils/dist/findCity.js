var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
import fetch from 'node-fetch';
const wait = (ms) => new Promise(resolve => setTimeout(resolve, ms));
const fetchWithRetry = (url_1, ...args_1) => __awaiter(void 0, [url_1, ...args_1], void 0, function* (url, maxRetries = 3, baseDelay = 1000) {
    for (let attempt = 0; attempt < maxRetries; attempt++) {
        try {
            return yield fetch(url);
        }
        catch (error) {
            if (attempt === maxRetries - 1)
                throw error;
            const delay = baseDelay * Math.pow(2, attempt);
            yield wait(delay);
        }
    }
    throw new Error('Max retries reached');
});
export const getCityFromCoordinates = (lat, lon) => __awaiter(void 0, void 0, void 0, function* () {
    const url = `https://nominatim.openstreetmap.org/reverse?lat=${lat}&lon=${lon}&format=json`;
    try {
        const response = yield fetchWithRetry(url);
        const data = yield response.json();
        if (data && data.address && data.address.city) {
            return data.address.city;
        }
        return null;
    }
    catch (error) {
        console.error('Error fetching data from Nominatim:', error);
        return null;
    }
});
