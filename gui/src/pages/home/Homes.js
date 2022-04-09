import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';

import './Homes.css';

export default function Homes() {
  const [homes, setHomes] = useState([]);
  const navigate = useNavigate();

  useEffect(() => {
    async function fn() {
      let token = localStorage.getItem('token');
      let headers = {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer ' + token
      };
      try {
        const response = await axios.get('/api/homes', {
          headers
        })
        const data = response.data;
        setHomes(data);
      } catch (err) {
        console.error('Cannot get homes');
      }
    }

    fn();
  }, []);

  function createNewHome() {
    navigate(`/main/homes/new`);
  }

  function showHomeDetails(home) {
    navigate(`/main/homes/${home.id}`, {state: {home}});
  }

  function editHome(home) {
    navigate(`/main/homes/${home.id}/edit`, {state: {home}});
  }

  async function deleteHome(home) {
    let token = localStorage.getItem('token');
    let headers = {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer ' + token
    };
    try {
      // delete
      await axios.delete(`/api/homes/${home.id}`, {
        headers
      })
      // get
      const response = await axios.get('/api/homes', {
        headers
      })
      const data = response.data;
      setHomes(data);
    } catch (err) {
      console.error('Cannot get homes');
    }
  }

  return (
    <div className="App">
      <h1>Homes</h1>
      <button onClick={() => createNewHome()}>+ add</button>
      {homes && homes.map(home => (
        <div className="home" key={home}>
          <p>{home.name} @ {home.location} <span onClick={() => showHomeDetails(home)}>view</span>&nbsp;
            <span onClick={() => editHome(home)}>edit</span>&nbsp;
            <span onClick={() => deleteHome(home)}>delete</span></p>
        </div>
      ))}
    </div>
  )
}

