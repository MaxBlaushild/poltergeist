export interface House {
  id: string;
  name: string;
}

export interface Secret {
  id: string;
  ordinal: number;
  body: string;
}

export type SubmissionStatus = 'submitted' | 'verified' | 'rejected';

export interface MissionSubmission {
  status: SubmissionStatus;
  playerAnswer: string;
  awardedBt: number;
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
  house?: House;
  // Gated — only present once content is unlocked.
  postAct1Context?: string;
  secrets?: Secret[];
  missions?: Mission[];
}

export interface GameState {
  currentAct: 'pre_event' | 'act1' | 'act2' | 'act3' | 'quiz' | 'resolved';
  contentUnlocked: boolean;
  quizOpen: boolean;
  activeNotificationId: string | null;
}

export interface Notification {
  id: string;
  title: string;
  body: string;
}

export interface QuizQuestion {
  id: string;
  ordinal: number;
  prompt: string;
  questionType: 'multiple_choice' | 'open';
  options: string[];
  answer: string;
}

export interface QuizResponse {
  quizOpen: boolean;
  submitted: boolean;
  questions: QuizQuestion[];
}

export interface MeResponse {
  player: { id: string; guestLabel: string };
  gameState: GameState;
  character: Character | null;
  notification: Notification | null;
}

export interface HouseStanding {
  houseId: string;
  name: string;
  sortOrder: number;
  favor: number;
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
  house: { id: string; name: string; favor: number };
  members: HouseMember[];
  log: HouseFavorLogEntry[];
}
