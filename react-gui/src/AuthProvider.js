import React, { useState } from 'react';
import AuthContext from './AuthContext';
import { setToken, removeToken, isLoggedIn } from './auth.util';

export default function AuthProvider({children}) {
  const [tokenState, setTokenState] = useState(null)

  const login = token => {
    setToken(token);
    setTokenState(token);
  };

  const logout = () => {
    removeToken();
    setTokenState(null);
  };

  const isLogged = () => {
    return isLoggedIn();
  };

  const value = {
    tokenState,
    isLogged,
    login,
    logout
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
}
