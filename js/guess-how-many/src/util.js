import { v4 as uuidv4 } from 'uuid';

export const getUserID = () => {
  const id = localStorage.getItem('user-id');
  if (id) {
    try {
      return {
        userId: id,
      };
    } catch (e) {
      localStorage.removeItem('user-id');
    }
  }

  let ephemeralUserId = localStorage.getItem('ephemeral-user-id');
  if (!ephemeralUserId) {
    ephemeralUserId = uuidv4();
    localStorage.setItem('ephemeral-user-id', ephemeralUserId);
  }

  return { ephemeralUserId };
};
