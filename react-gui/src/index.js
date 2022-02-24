import React from 'react';
import {render} from 'react-dom'
import axios from 'axios';

import './index.css';
import App from './App';
import reportWebVitals from './reportWebVitals';
import {removeToken} from "./auth/auth-utils";

axios.defaults.baseURL = process.env.REACT_APP_BASEURL;

const responseErrorHandler = error => {
  if (error.response.status === 401) {
    console.log('responseErrorHandler - 401');
    removeToken();
    window.location.href = process.env.REACT_APP_BASEURL;
    // Add your logic to
    //  1. Redirect user to LOGIN
    //  2. Reset authentication from localstorage/sessionstorage
  }
  return Promise.reject(error);
}
axios.interceptors.response.use(
  response => response,
  error => responseErrorHandler(error)
);

render(
  <React.StrictMode>
    <App/>
  </React.StrictMode>,
  document.getElementById('root')
);

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
