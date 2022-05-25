import { removeToken } from '../auth/auth-utils';

export const getApi = (url) => request('GET', url, null);
export const postApi = (url, body) => request('POST', url, body);
export const putApi = (url, body) => request('PUT', url, body);
export const deleteApi = (url) => request('DELETE', url, null);

function request(method, url, body) {
  const requestOptions = {
    method,
    headers: getHeaders()
  };
  if (body) {
    requestOptions.body = JSON.stringify(body);
  }
  return fetch(url, requestOptions).then(handleResponse);
}

function getHeaders() {
  let token = localStorage.getItem('token');
  return {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer ' + token
  };
}

function handleResponse(response) {
  return response.text().then(text => {
    let json;
    try {
      json = JSON.parse(text);
    } catch (err) {
      return Promise.reject(new Error(`API json parsing error: cannot parse text: ${text}`));
    }

    if (!response.ok) {
      if (response.status === 401 || response.status === 403) {
        console.log('response.status - 401 or 403 - logging out');
        //  1. Redirect user to LOGIN
        //  2. Reset authentication from localstorage/sessionstorage
        removeToken();
        window.location.href = '/';
      }
      return Promise.reject(new Error(`API error: ${text}`));
    }
    console.log('API success - json result:', json);
    return json;
  });
}
