export interface House {
  id: string;
  name: string;
  tagline?: string;
}

export interface Secret {
  id: string;
  ordinal: number;
  body: string;
}

export type SubmissionStatus = 'submitted' | 'approved' | 'redeemed' | 'rejected';

export interface MissionSubmission {
  status: SubmissionStatus;
  playerAnswer: string;
  awardedBt: number;
  photoIds: string[];
}

export interface Mission {
  id: string;
  ordinal: number;
  tier: 'easy' | 'medium' | 'hard';
  rewardBt: number;
  prompt: string;
  answerFormat: string;
  submission?: MissionSubmission | null;
}

export interface Character {
  id: string;
  name: string;
  title: string;
  roleType: string;
  preEventInfo: string;
  imageUrl?: string;
  house?: House;
  // Gated — only present once content is unlocked.
  postAct1Context?: string;
  secrets?: Secret[];
  missions?: Mission[];
}

export interface GameState {
  currentAct: 'pre_event' | 'act1' | 'act2' | 'act3' | 'quiz' | 'quiz_part1' | 'quiz_part2' | 'resolved';
  contentUnlocked: boolean;
  quizPart1Open: boolean;
  quizPart2Open: boolean;
  quizPart1OpenedAt: string | null;
  activeNotificationId: string | null;
}

export interface Notification {
  id: string;
  title: string;
  body: string;
}

export interface QuizPart1 {
  open: boolean;
  openedAt: string | null;
  prompt: string;
  submitted: boolean;
  answer: string;
}

export interface QuizPart2Question {
  id: string;
  ordinal: number;
  prompt: string;
  tier: string;
  type?: string; // 'multiple_choice' (default) | 'number'
  options: string[];
  answer: string;
}

export interface QuizPart2 {
  open: boolean;
  submitted: boolean;
  // Sequential flow: only the current (first-unanswered) question is sent.
  total?: number;
  answered?: number;
  questions: QuizPart2Question[];
}

export interface QuizResponse {
  part1: QuizPart1;
  part2: QuizPart2;
}

export interface GameWinner {
  characterId: string;
  characterName: string;
  house?: string;
}

export interface Game {
  id: string;
  ordinal: number;
  name: string;
  status: 'pending' | 'played';
  first: GameWinner | null;
  second: GameWinner | null;
  third: GameWinner | null;
}

export interface MeResponse {
  // guestLabel is a GM-only roster field (the real player name) and is not sent here.
  player: { id: string };
  gameState: GameState;
  character: Character | null;
  notification: Notification | null;
}

export interface HouseStanding {
  houseId: string;
  name: string;
  sortOrder: number;
  favor: number; // ledger base (excludes item effects)
  itemFavor?: number; // live "+X" overlay from owned items
}

export interface HouseMember {
  id: string;
  name: string;
  title: string;
}

export interface HouseFavorLogEntry {
  id: string;
  delta: number;
  reason: string;
  gmName: string;
  source: string;
  createdAt: string;
}

export interface HouseOverview {
  house: { id: string; name: string; favor: number; itemFavor?: number };
  members: HouseMember[];
  log: HouseFavorLogEntry[];
}
