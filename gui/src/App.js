import React from 'react';
import { BrowserRouter as Router, Route, Routes } from 'react-router-dom';

import { CssBaseline, ThemeProvider, createTheme } from '@mui/material';

import AuthProvider from './auth/AuthProvider';
import RequireAuth from './auth/RequireAuth';

import Login from './components/Login';
import PostLogin from './components/PostLogin';
import Main from './components/Main';
import Homes from './components/home/Homes';
import Devices from './components/devices/Devices';
import Profile from './components/Profile';
import EditHome from './components/home/EditHome';
import DeviceSettings from './components/devices/DeviceSettings';
import Controller from './components/devices/Controller';
import Sensor from './components/devices/Sensor';

const darkTheme = createTheme({
  palette: {
    mode: 'dark'
  },
});

export default function App() {
  return (
    <ThemeProvider theme={darkTheme}>
      {/* CssBaseline kickstart an elegant, consistent, and simple baseline to build upon. */}
      <CssBaseline />
      <AuthProvider>
        <Router>
          <Routes>
            <Route path="/" element={<Login/>}/>
            <Route path="login" element={<Login/>}/>
            <Route path="postlogin" element={<PostLogin/>}/>
            <Route path="profile" element={<RequireAuth> <Profile/> </RequireAuth>}/>
            <Route path="main" element={<Main/>}>
              <Route index element={<RequireAuth><Devices/></RequireAuth>}/>
              <Route index path="devices" element={<RequireAuth><Devices/></RequireAuth>}/>
              <Route path="devices/:id" element={<RequireAuth><DeviceSettings/></RequireAuth>}/>
              <Route path="devices/:id/controller" element={<RequireAuth><Controller/></RequireAuth>}/>
              <Route path="devices/:id/sensor" element={<RequireAuth><Sensor/></RequireAuth>}/>
              <Route path="homes" element={<RequireAuth> <Homes/> </RequireAuth>}/>
              <Route path="homes/:id/edit" element={<RequireAuth> <EditHome/> </RequireAuth>}/>
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
    </ThemeProvider>
  )
}
