// token name used in local storage
export const TOKEN_NAME = 'token';

export function isLoggedIn(tokenName = TOKEN_NAME) {
  return !!localStorage.getItem(tokenName);
}

export function getToken(tokenName = TOKEN_NAME) {
  return localStorage.getItem(tokenName);
}

export function setToken(token, tokenName = TOKEN_NAME) {
  localStorage.setItem(tokenName, token);
}

export function removeToken(tokenName = TOKEN_NAME) {
  localStorage.removeItem(tokenName);
}
