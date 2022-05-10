import React from 'react';
import {render} from 'react-dom'

import './index.css';
import App from './App';
import reportWebVitals from './reportWebVitals';

// const responseErrorHandler = error => {
//   if (error.response.status === 401) {
//     console.log('responseErrorHandler - 401');
//     removeToken();
//     window.location.href = '/';
//     // Add your logic to
//     //  1. Redirect user to LOGIN
//     //  2. Reset authentication from localstorage/sessionstorage
//   }
//   return Promise.reject(error);
// }

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
