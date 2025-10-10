import React, { createContext, ReactNode, useContext, useState } from 'react';
import { useUser } from '@poltergeist/hooks';
import { User } from '@poltergeist/types';

interface UserContextType {
  user: User | null;
  loading: boolean;
  error: Error | null;
  username: string | null;
  setUsername: (username: string) => void;
}

export const UserContext = createContext<UserContextType | null>(null);

export const UserContextProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [username, setUsername] = useState<string | null>(null);
  const { user, loading, error } = useUser(username); 

  return <UserContext.Provider value={{ 
    user, 
    loading, 
    error, 
    username, 
    setUsername 
  }}>
    {children}
  </UserContext.Provider>;
};

export const useUserContext = () => {
  return useContext(UserContext);
};