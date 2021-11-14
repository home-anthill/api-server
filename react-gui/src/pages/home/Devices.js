import React, { useEffect, useState } from 'react';
import axios from 'axios';

import './Devices.css';

export default function Devices() {
  const [devices, setDevices] = useState([]);

  useEffect(() => {
    async function fn() {
      let token = localStorage.getItem('token');
      let headers = {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer ' + token
      };

      const response = await axios.get('http://localhost:8082/api/airconditioners', {
        headers
      })
      const data = response.data;
      console.log('Devices: ', data);
      setDevices(data);
    }
    fn();
  }, []);

  return (
    <div className="App">
      <h1>Devices</h1>
      {devices.map(device => (
        <div className="device" key={device}>
          <p>{device.name} - {device.manufacturer} - {device.model}</p>
        </div>
      ))}
    </div>
  )
}

