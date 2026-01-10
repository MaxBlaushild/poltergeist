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
    allGreensAchieved: boolean;
    allPurplesAchieved: boolean;
}
export interface ButtonConfig {
    slot: number;
    hueLightId?: number | null;
    baseHue: number;
}
export type PuzzleColor = 'Off' | 'Blue' | 'Red' | 'White' | 'Yellow' | 'Purple' | 'Green';
export declare const PUZZLE_COLORS: PuzzleColor[];
export declare const COLOR_TO_INDEX: Record<PuzzleColor, number>;
export declare const INDEX_TO_COLOR: Record<number, PuzzleColor>;
