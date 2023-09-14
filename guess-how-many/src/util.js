import { v4 as uuidv4 } from 'uuid';

export const getUserID = () => {
  let userId = localStorage.getItem('user-id');

  if (userId) {
    return userId;
  }

  userId = uuidv4();
  localStorage.setItem('user-id', userId);
  return userId;
};
