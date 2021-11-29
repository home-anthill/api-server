import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';

import './Devices.css';

export default function Devices() {
  const [devices, setDevices] = useState([]);
  const navigate = useNavigate();

  useEffect(() => {
    async function fn() {
      const token = localStorage.getItem('token');
      const headers = {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer ' + token
      };
      try {
        const response = await axios.get('http://localhost:8082/api/devices', {
          headers
        })
        const data = response.data;
        console.log('Devices: ', data);
        setDevices(data);
      } catch (err) {
        console.error('Cannot get devices');
      }
    }

    fn();
  }, []);

  function showDeviceDetails(device) {
    navigate(`/main/devices/${device.id}`, {state: {device}});
  }

  return (
    <div className="App">
      <h1>Devices</h1>
      {devices.map(device => (
        <div className="device" key={device}>
          <p onClick={() => showDeviceDetails(device)}>{device.name} - {device.manufacturer} - {device.model}</p>
        </div>
      ))}
    </div>
  )
}

