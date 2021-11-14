import React, { useEffect, useState } from 'react';

import logo from '../logo.svg';
import { isLoggedIn } from '../auth/auth-utils';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';

export default function Login() {
  const [state, setState] = useState({loginURL: null});
  const navigate = useNavigate();

  function login() {
    window.location.href = state.loginURL;
  }

  useEffect(() => {
    async function fn() {
      if (isLoggedIn()) {
        console.log('Already logged in');
        navigate('/main');
      } else {
        console.log('getting login URL');
        try {
          const response = await axios.get('http://localhost:8082/api/login');
          const data = response.data;
          console.log('responseData', data);
          if (data) {
            const loginURL = data.loginURL;
            console.log('loginURL found:', loginURL)
            setState({loginURL: loginURL});
          }
        } catch (err) {
          console.error('Cannot login', err);
        }
      }
    }
    fn();
  }, [])

  return (
    <div className="App">
      <h1>Login</h1>
      <img src={logo} className="App-logo" alt="logo"/>
      <button onClick={login}>Login</button>
      <div>{state.loginURL}</div>
    </div>
  )
}

