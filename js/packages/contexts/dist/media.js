var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
import { jsx as _jsx } from "react/jsx-runtime";
import { createContext, useContext, useCallback } from 'react';
import * as React from 'react';
import APIClient from '@poltergeist/api-client';
export const MediaContext = createContext(undefined);
export const useMediaContext = () => {
    const context = useContext(MediaContext);
    if (!context) {
        throw new Error('useMediaContext must be used within a MediaContextProvider');
    }
    return context;
};
export const MediaContextProvider = ({ children }) => {
    const apiClient = new APIClient(process.env.REACT_APP_API_URL || '');
    const getPresignedUploadURL = React.useCallback((bucket, key) => __awaiter(void 0, void 0, void 0, function* () {
        try {
            const response = yield apiClient.post(`/sonar/media/uploadUrl`, { bucket, key });
            return response.url;
        }
        catch (error) {
            console.error('Failed to get presigned upload URL', error);
            return null;
        }
    }), [apiClient]);
    const openCameraAndCaptureImage = useCallback(() => __awaiter(void 0, void 0, void 0, function* () {
        try {
            const mediaStream = yield navigator.mediaDevices.getUserMedia({ video: true });
            const video = document.createElement('video');
            video.srcObject = mediaStream;
            video.play();
            return new Promise((resolve, reject) => {
                video.addEventListener('canplay', () => {
                    const canvas = document.createElement('canvas');
                    canvas.width = video.videoWidth;
                    canvas.height = video.videoHeight;
                    const context = canvas.getContext('2d');
                    context === null || context === void 0 ? void 0 : context.drawImage(video, 0, 0, canvas.width, canvas.height);
                    canvas.toBlob(blob => {
                        if (blob) {
                            resolve(blob);
                        }
                        else {
                            reject(new Error('Failed to convert canvas to blob'));
                        }
                        mediaStream.getTracks().forEach(track => track.stop());
                        video.pause();
                    }, 'image/webp');
                });
            });
        }
        catch (error) {
            console.error('Failed to open camera or capture image', error);
            return null;
        }
    }), []);
    const uploadMedia = useCallback((url, file) => __awaiter(void 0, void 0, void 0, function* () {
        try {
            const response = yield fetch(url, {
                method: 'PUT',
                body: file,
                headers: {
                    'Content-Type': file.type,
                },
            });
            return response.ok;
        }
        catch (error) {
            console.error('Failed to upload media', error);
            return false;
        }
    }), []);
    const uploadImage = useCallback((key, image) => __awaiter(void 0, void 0, void 0, function* () {
        const presignedURL = yield getPresignedUploadURL('crew-challenge-images', key);
        if (!presignedURL)
            return;
        yield uploadMedia(presignedURL, image);
    }), [getPresignedUploadURL, uploadMedia]);
    return (_jsx(MediaContext.Provider, Object.assign({ value: { getPresignedUploadURL, uploadMedia, openCameraAndCaptureImage } }, { children: children })));
};
