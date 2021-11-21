import React, { useEffect, useState } from 'react';
import axios from 'axios';

import './Devices.css';

export default function Devices() {
  const [devices, setDevices] = useState([]);
  const [homes, setHomes] = useState([]);

  useEffect(() => {
    async function fn() {
      let token = localStorage.getItem('token');
      let headers = {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer ' + token
      };

      let response = await axios.get('http://localhost:8082/api/airconditioners', {
        headers
      })
      let data = response.data;
      console.log('Devices: ', data);
      setDevices(data);

      response = await axios.get('http://localhost:8082/api/homes', {
        headers
      })
      data = response.data;
      console.log('Homes: ', data);
      setHomes([{name: '---', rooms: []}, ...data]);
    }
    fn();
  }, []);

  function onChangeHome() {

  }

  return (
    <div className="App">
      <h1>Devices</h1>
      {devices.map(device => (
        <div className="device" key={device}>
          <p>{device.name} - {device.manufacturer} - {device.model}</p>
          <select name="home" id="homes" onChange={e => onChangeHome()}>
            {
              homes.map((home, index) => <option key={index} value={home.id}>{home.name}</option>)
            }
          </select>
          <select name="home" id="homes" onChange={e => onChangeHome()}>
            {
              homes.map((home, index) => <option key={index} value={home.id}>{home.name}</option>)
            }
          </select>

          {home.rooms?.map(room => (
            <div className="room" key={room} onClick={() => showRoomDetails(home, room)}>
              <p>{room.name} @ {room.floor}</p>
            </div>
          ))}
        </div>
      ))}
    </div>
  )
}

