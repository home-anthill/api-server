import React from 'react';
import Navbar from '../components/Navbar';
import { Link, Outlet } from 'react-router-dom';

export default function Main() {
  return (
    <div>
      <Navbar/>
      <main>
        <Link to="/main/homes">Homes</Link>
        <Link to="/main/devices">Devices</Link>
        <Outlet/>
      </main>
    </div>
  )
}
