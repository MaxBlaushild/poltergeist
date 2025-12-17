export interface UtilityClosetPuzzle {
  id: string;
  createdAt: string;
  updatedAt: string;
  button0HueLightId?: number | null;
  button1HueLightId?: number | null;
  button2HueLightId?: number | null;
  button3HueLightId?: number | null;
  button4HueLightId?: number | null;
  button5HueLightId?: number | null;
  button0CurrentHue: number;
  button1CurrentHue: number;
  button2CurrentHue: number;
  button3CurrentHue: number;
  button4CurrentHue: number;
  button5CurrentHue: number;
  button0BaseHue: number;
  button1BaseHue: number;
  button2BaseHue: number;
  button3BaseHue: number;
  button4BaseHue: number;
  button5BaseHue: number;
}

export interface ButtonConfig {
  slot: number;
  hueLightId?: number | null;
  baseHue: number;
}

export type PuzzleColor = 'Off' | 'Blue' | 'Green' | 'White' | 'Red' | 'Purple' | 'Gold';

export const PUZZLE_COLORS: PuzzleColor[] = ['Off', 'Blue', 'Green', 'White', 'Red', 'Purple', 'Gold'];

export const COLOR_TO_INDEX: Record<PuzzleColor, number> = {
  'Off': 0,
  'Blue': 1,
  'Green': 2,
  'White': 3,
  'Red': 4,
  'Purple': 5,
  'Gold': 6,
};

export const INDEX_TO_COLOR: Record<number, PuzzleColor> = {
  0: 'Off',
  1: 'Blue',
  2: 'Green',
  3: 'White',
  4: 'Red',
  5: 'Purple',
  6: 'Gold',
};
