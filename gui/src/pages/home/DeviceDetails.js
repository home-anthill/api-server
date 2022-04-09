import React, { useEffect, useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import axios from 'axios';

import './Devices.css';
import Values from '../../components/Values';

const DEFAULT_HOME = {name: '---', rooms: []};
const DEFAULT_ROOM = {name: '---'};

export default function DeviceDetails() {
  const {state} = useLocation();
  const device = state.device;
  const navigate = useNavigate();

  const [homes, setHomes] = useState([]);
  const [rooms, setRooms] = useState([]);

  const [selectedHome, setSelectedHome] = useState(DEFAULT_HOME);
  const [selectedRoom, setSelectedRoom] = useState(DEFAULT_ROOM);

  useEffect(() => {
    async function fn() {
      console.log('useEffect 1 - fn');
      const token = localStorage.getItem('token');
      const headers = {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer ' + token
      };
      try {
        const response = await axios.get('/api/homes', {
          headers
        })
        const data = response.data;
        console.log('Homes: ', data);
        const homes = [DEFAULT_HOME, ...data];
        setHomes(homes);

        let homeFound;
        let roomFound;
        homes.forEach(home => {
          home.rooms.forEach(room => {
            if (room && room.devices && room.devices.find(dev => dev === device.id)) {
              homeFound = home;
              roomFound = room;
            }
          });
        });

        console.log('Init: homeFound: ', homeFound);

        if (homeFound) {
          setSelectedHome(homeFound);
          setRooms([DEFAULT_ROOM, ...homeFound.rooms])
          setSelectedRoom(roomFound);
        }
      } catch (err) {
        console.error('Cannot get homes');
      }
    }

    fn();
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
    console.log('onSave', {selectedHome, selectedRoom});
    const token = localStorage.getItem('token');
    const headers = {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer ' + token
    };
    const newRoom = Object.assign({}, selectedRoom);
    if (!newRoom.devices) {
      newRoom.devices = [device.id];
    } else {
      newRoom.devices.push(device.id);
    }
    try {
      await axios.put(`/api/homes/${selectedHome.id}/rooms/${selectedRoom.id}`,
        newRoom,
        {
          headers
        }
      );
      // navigate back
      navigate(-1);
    } catch (err) {
      console.error('Cannot save device assigning it to this room');
    }
  }

  async function onRemove() {
    console.log('onRemove');
    const token = localStorage.getItem('token');
    const headers = {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer ' + token
    };
    try {
      await axios.delete(`/api/devices/${device.id}?homeId=${selectedHome.id}&roomId=${selectedRoom.id}`,
        {
          headers
        }
      );
      // navigate back
      navigate(-1);
    } catch (err) {
      console.error('Cannot remove device');
    }
  }

  return (
    <div className="App">
      <h1>Device</h1>
      <p>{device.name} - {device.manufacturer} - {device.model}</p>
      <br/>
      <select name="home" id="homes" onChange={event => onChangeHome(event)} value={selectedHome.id}>
        {
          homes.map(home => <option key={home.id} value={home.id}>{home.name}</option>)
        }
      </select>
      <select name="room" id="rooms" onChange={event => onChangeRoom(event)} value={selectedRoom.id}>
        {
          rooms.map(room => <option key={room.id} value={room.id}>{room.name}</option>)
        }
      </select>
      <br/>
      {selectedHome.name !== DEFAULT_HOME.name &&
      selectedRoom.name !== DEFAULT_ROOM.name &&
      <button onClick={() => onSave()}>Save</button>
      }
      <br/>
      <br/>
      <button onClick={() => onRemove()}>Remove this Device</button>
      <br/>
      <br/>
      <Values device={device}/>
    </div>
  )
}

