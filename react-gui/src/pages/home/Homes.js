import React, { useEffect, useState } from 'react';
import axios from 'axios';
import { useNavigate } from 'react-router-dom';

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

      const response = await axios.get('http://localhost:8082/api/homes', {
        headers
      })
      const data = response.data;
      console.log('Homes: ', data);
      setHomes(data);
    }

    fn();
  }, []);

  function showHomeDetails(home) {
    navigate(`/main/homes/${home.id}`, {state: {home}});
  }

  return (
    <div className="App">
      <h1>Homes</h1>
      {homes && homes.map(home => (
        <div className="home" key={home} onClick={() => showHomeDetails(home)}>
          <p>{home.name} @ {home.location}</p>
        </div>
      ))}
    </div>
  )
}

