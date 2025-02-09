import { createContext, useContext, useState, useCallback } from 'react';
import * as React from 'react';
import { useAPI } from './api';
import APIClient from '@poltergeist/api-client';

export type MediaContextType = {
  getPresignedUploadURL: (bucket: string, key: string) => Promise<string | null>;
  uploadMedia: (url: string, file: Blob) => Promise<boolean>;
  openCameraAndCaptureImage: () => Promise<Blob | null>;
};

export const MediaContext = createContext<MediaContextType | undefined>(undefined);

export const useMediaContext = () => {
  const context = useContext(MediaContext);
  if (!context) {
    throw new Error('useMediaContext must be used within a MediaContextProvider');
  }
  return context;
};

export const MediaContextProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const apiClient = new APIClient(process.env.REACT_APP_API_URL || '');

  const getPresignedUploadURL = React.useCallback(async (bucket: string, key: string): Promise<string | null> => {
    try {
      const response = await apiClient.post<{ url: string }>(`/sonar/media/uploadUrl`, { bucket, key });
      return response.url;
    } catch (error) {
      console.error('Failed to get presigned upload URL', error);
      return null;
    }
  }, [apiClient]);

  const openCameraAndCaptureImage = useCallback(async (): Promise<Blob | null> => {
    try {
      const mediaStream = await navigator.mediaDevices.getUserMedia({ video: true });
      const video = document.createElement('video');
      video.srcObject = mediaStream;
      video.play();

      return new Promise((resolve, reject) => {
        video.addEventListener('canplay', () => {
          const canvas = document.createElement('canvas');
          canvas.width = video.videoWidth;
          canvas.height = video.videoHeight;
          const context = canvas.getContext('2d');
          context?.drawImage(video, 0, 0, canvas.width, canvas.height);
          canvas.toBlob(blob => {
            if (blob) {
              resolve(blob);
            } else {
              reject(new Error('Failed to convert canvas to blob'));
            }
            mediaStream.getTracks().forEach(track => track.stop());
            video.pause();
          }, 'image/webp');
        });
      });
    } catch (error) {
      console.error('Failed to open camera or capture image', error);
      return null;
    }
  }, []);

  const uploadMedia = useCallback(async (url: string, file: Blob): Promise<boolean> => {
    try {

      const response = await fetch(url, {
        method: 'PUT',
        body: file,
        headers: {
          'Content-Type': file.type,
        },
      });

      return response.ok;
    } catch (error) {
      console.error('Failed to upload media', error);
      return false;
    }
  }, []);

  const uploadImage = useCallback(async (key: string, image: Blob) => {
    const presignedURL = await getPresignedUploadURL('crew-challenge-images', key);
    if (!presignedURL) return;
    await uploadMedia(presignedURL, image);
  }, [getPresignedUploadURL, uploadMedia]);

  return (
    <MediaContext.Provider value={{ getPresignedUploadURL, uploadMedia, openCameraAndCaptureImage }}>
      {children}
    </MediaContext.Provider>
  );
};
