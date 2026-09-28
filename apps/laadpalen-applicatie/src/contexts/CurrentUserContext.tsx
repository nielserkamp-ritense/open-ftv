import React, { createContext, useContext, useState } from "react";

interface CurrentUserContextType {
  currentUser: string;
  setCurrentUser: (user: string) => void;
}

const CurrentUserContext = createContext<CurrentUserContextType | undefined>(undefined);

export const CurrentUserProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [currentUser, setCurrentUserState] = useState<string>(() => {
    return localStorage.getItem("currentUser") || "Rick";
  });

  const setCurrentUser = (user: string) => {
    setCurrentUserState(user);
    localStorage.setItem("currentUser", user);
  };

  return (
    <CurrentUserContext.Provider value={{ currentUser, setCurrentUser }}>
      {children}
    </CurrentUserContext.Provider>
  );
};

export const useCurrentUser = () => {
  const context = useContext(CurrentUserContext);
  if (!context) {
    throw new Error("useCurrentUser must be used within CurrentUserProvider");
  }
  return context;
};
