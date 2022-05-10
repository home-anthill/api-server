import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';

import { Button, Link, Typography } from '@mui/material';

import './Login.css';
import { isLoggedIn } from '../auth/auth-utils';
import logoPng from '../air-conditioner.png'

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
          const response = await fetch('/api/login');
          const body = await response.json();
          console.log('responseData', body);
          if (body) {
            const loginURL = body.loginURL;
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
    <div className="Login">
      <Typography variant="h2" component="h1">
        Welcome to air-conditioner
      </Typography>
      <img className="Logo" src={logoPng} width="250" height="auto" alt="Air conditioner"></img>
      <Button variant="contained" className="BtnContained" onClick={login} disabled={!state.loginURL}>LOGIN</Button>
      <Link href="https://www.flaticon.com/free-icons/air-conditioner"
            sx={{marginTop: '45px'}}
            underline="hover"
            title="air conditioner icons">
        Air conditioner icons created by Freepik - Flaticon
      </Link>
    </div>
  )
}

