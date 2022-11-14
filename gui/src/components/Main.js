import React from 'react';
import Navbar from '../shared/Navbar';
import { Outlet } from 'react-router-dom';

export default function Main() {
  return (
    <div>
      <Navbar/>
      <main>
        <Outlet/>
      </main>
    </div>
  )
}
