import React from 'react';
import { BrowserRouter as Router, Route, Routes } from 'react-router-dom';

import './App.css';

import Login from './pages/Login';
import PostLogin from './pages/PostLogin';
import Main from './pages/Main';
import Homes from './pages/home/Homes';
import Devices from './pages/home/Devices';
import AuthProvider from './AuthProvider';
import RequireAuth from './RequireAuth';

import axios from "axios";
import { removeToken } from './auth.util';

const responseSuccessHandler = response => {
  return response;
};
const responseErrorHandler = error => {
  if (error.response.status === 401) {
    console.log('responseErrorHandler - 401');
    removeToken();
    // Add your logic to
    //  1. Redirect user to LOGIN
    //  2. Reset authentication from localstorage/sessionstorage
  }
  return Promise.reject(error);
}
axios.interceptors.response.use(
  response => responseSuccessHandler(response),
  error => responseErrorHandler(error)
);

export default function App() {
  return (
    <AuthProvider>
      <Router>
        <Routes>
          <Route path="/" element={<Login/>}/>
          <Route path="login" element={<Login/>}/>
          <Route path="postlogin" element={<PostLogin/>}/>
          <Route path="main" element={<Main/>}>
            <Route index element={<RequireAuth> <Homes/> </RequireAuth>}/>
            <Route index path="homes" element={<RequireAuth> <Homes/> </RequireAuth>}/>
            <Route path="devices" element={<RequireAuth><Devices/></RequireAuth>}/>
          </Route>
          <Route
            path="*"
            element={
              <main style={{padding: '1rem'}}>
                <p>There's nothing here!</p>
              </main>
            }
          />
        </Routes>
      </Router>
    </AuthProvider>
  )
}
