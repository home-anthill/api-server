import React, { useEffect } from 'react';
import Navbar from '../components/Navbar';
import { Outlet } from 'react-router-dom';

export default function Main() {

  // useEffect(() => {
  //   let token = localStorage.getItem('token');
  //   let headers = {
  //     'Content-Type': 'application/json',
  //     'Authorization': 'Bearer ' + token
  //   };
  //
  //   fetch('http://localhost:8082/api/homes', {headers})
  //     .then(response => response.json())
  //     .then(responseData => {
  //       console.log('responseData', responseData);
  //     });
  // });

  return (
    <div>
      <Navbar/>
      <main>
        <Outlet/>
      </main>
    </div>
  )
}
