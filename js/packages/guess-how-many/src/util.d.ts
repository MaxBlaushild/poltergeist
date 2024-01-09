export declare const getUserID: () =>
  | {
      userId: string;
      ephemeralUserId?: undefined;
    }
  | {
      ephemeralUserId: string | null;
      userId?: undefined;
    };
