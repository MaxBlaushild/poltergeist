import { Point } from "./point";

export type Zone = {
  id: string;
  name: string;
  latitude: number;
  longitude: number;
  radius: number;
  createdAt: string;
  updatedAt: string;
  boundary: number[][];
  boundaryCoords: {
    latitude: number;
    longitude: number;
  }[];
  points: Point[];
};