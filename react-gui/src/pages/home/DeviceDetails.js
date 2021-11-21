import React, { useEffect, useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import axios from 'axios';

import './Devices.css';

const DEFAULT_HOME = {name: '---', rooms: []};
const DEFAULT_ROOM = {name: '---'};

export default function DeviceDetails() {
  const {state} = useLocation();
  const navigate = useNavigate();
  const device = state.device;

  const [homes, setHomes] = useState([]);
  const [rooms, setRooms] = useState([]);

  const [selectedHome, setSelectedHome] = useState(DEFAULT_HOME);
  const [selectedRoom, setSelectedRoom] = useState(DEFAULT_ROOM);

  useEffect(() => {
    async function fn() {
      const token = localStorage.getItem('token');
      const headers = {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer ' + token
      };
      try {
        const response = await axios.get('http://localhost:8082/api/homes', {
          headers
        })
        const data = response.data;
        console.log('Homes: ', data);
        setHomes([DEFAULT_HOME, ...data]);
      } catch (err) {
        console.error('Cannot get homes');
      }
    }

    fn();
    setSelectedHome(DEFAULT_HOME);
    setSelectedRoom(DEFAULT_ROOM);
  }, []);

  function onChangeHome(event) {
    const index = event.target.selectedIndex;
    if (index === 0) {
      setSelectedHome(DEFAULT_HOME);
      return;
    }
    const home = homes[index];
    console.log('onChangeHome - home: ', home);
    setSelectedHome(home);
    setRooms([DEFAULT_ROOM, ...home.rooms])
  }

  function onChangeRoom(event) {
    const index = event.target.selectedIndex;
    if (index === 0) {
      setSelectedRoom(DEFAULT_ROOM);
      return;
    }
    const room = rooms[index];
    console.log('onChangeRoom - room: ', room);
    setSelectedRoom(room);
  }

  async function onSave() {
    console.log('Saving with selectedHome and selectedRoom', {selectedHome, selectedRoom});
    const token = localStorage.getItem('token');
    const headers = {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer ' + token
    };
    const newRoom = Object.assign({}, selectedRoom);
    if (!newRoom.airConditioners) {
      newRoom.airConditioners = [device.id];
    } else {
      newRoom.airConditioners.push(device.id);
    }
    try {
      await axios.put(`http://localhost:8082/api/homes/${selectedHome.id}/rooms/${selectedRoom.id}`,
        newRoom,
        {
          headers
        }
      );
    } catch (err) {
      console.error('Cannot put room with paired device');
    }
  }

  return (
    <div className="App">
      <h1>Device</h1>
      <p>{device.name} - {device.manufacturer} - {device.model}</p>
      <br/>
      <select name="home" id="homes" onChange={event => onChangeHome(event)}>
        {
          homes.map(home => <option key={home.id} value={home.id}>{home.name}</option>)
        }
      </select>
      <br/>
      <select name="room" id="rooms" onChange={event => onChangeRoom(event)}>
        {
          rooms.map(room => <option key={room.id} value={room.id}>{room.name}</option>)
        }
      </select>
      <br/>
      {selectedHome.name !== DEFAULT_HOME.name &&
      selectedRoom.name !== DEFAULT_ROOM.name &&
      <button onClick={() => onSave()}>Save</button>
      }
    </div>
  )
}

