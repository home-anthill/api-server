import React from 'react';
import { BrowserRouter as Router, Route, Routes } from 'react-router-dom';
import axios from "axios";

import './App.css';

import AuthProvider from './auth/AuthProvider';
import RequireAuth from './auth/RequireAuth';

import Login from './pages/Login';
import PostLogin from './pages/PostLogin';
import Main from './pages/Main';
import Homes from './pages/home/Homes';
import Devices from './pages/home/Devices';
import HomeDetails from './pages/home/HomeDetails';
import RoomDetails from './pages/home/RoomDetails';
import Profile from './pages/Profile';
import NewHome from './pages/home/NewHome';
import EditHome from './pages/home/EditHome';
import DeviceDetails from './pages/home/DeviceDetails';

export default function App() {
  return (
    <AuthProvider>
      <Router>
        <Routes>
          <Route path="/" element={<Login/>}/>
          <Route path="login" element={<Login/>}/>
          <Route path="postlogin" element={<PostLogin/>}/>
          <Route path="profile" element={<RequireAuth> <Profile/> </RequireAuth>}/>
          <Route path="main" element={<Main/>}>
            <Route index element={<RequireAuth> <Homes/> </RequireAuth>}/>
            <Route index path="homes" element={<RequireAuth> <Homes/> </RequireAuth>}/>
            <Route path="homes/new" element={<RequireAuth> <NewHome/> </RequireAuth>}/>
            <Route path="homes/:id" element={<RequireAuth> <HomeDetails/> </RequireAuth>}/>
            <Route path="homes/:id/edit" element={<RequireAuth> <EditHome/> </RequireAuth>}/>
            <Route path="homes/:id/rooms/:rid" element={<RequireAuth> <RoomDetails/> </RequireAuth>}/>
            <Route path="devices" element={<RequireAuth><Devices/></RequireAuth>}/>
            <Route path="devices/:id" element={<RequireAuth><DeviceDetails/></RequireAuth>}/>
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
