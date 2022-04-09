import React, { useState } from 'react';
import axios from 'axios';
import { useNavigate } from 'react-router-dom';

export default function NewHome() {
  const [home, setHome] = useState({name: '', location: ''});
  const [rooms, setRooms] = useState([]);
  const navigate = useNavigate();

  const submit = async e => {
    e.preventDefault(); // prevent default submit

    let token = localStorage.getItem('token');
    let headers = {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer ' + token
    };
    try {
      await
        axios.post(`/api/homes`, {
        name: home.name,
        location: home.location,
        rooms: rooms
      }, {
        headers
      });
      // navigate back
      navigate(-1);
    } catch (err) {
      console.error('Cannot add a new home');
    }
  }

  function onAddRoom() {
    console.log('add room');
    setRooms([...rooms, {name: '', floor: 0}]);
  }

  function onChangeRoom(index, newRoom) {
    const updated = [...rooms];
    updated[index] = newRoom;
    setRooms(updated)
  }

  return (
    <div className="App">
      <h1>New Home</h1>

      <form>
        <input
          value={home.name}
          onChange={event => setHome(Object.assign({}, home, {name: event.target.value}))}
          type="text"
          placeholder="Name"
          required
        />
        <input
          value={home.location}
          onChange={event => setHome(Object.assign({}, home, {location: event.target.value}))}
          type="text"
          placeholder="Location"
          required
        />
      </form>

      <p>Rooms</p>
      {rooms.map((room, index) => (
        <>
          <input
            value={room.name}
            onChange={event => onChangeRoom(index, {name: event.target.value, floor: room.floor})}
            type="text"
            placeholder="Room name"
            required
          />
          <input
            value={room.floor}
            onChange={event => onChangeRoom(index, {name: room.name, floor: +event.target.value})}
            type="number"
            placeholder="Room floor"
            required
          />
        </>
      ))}
      <button onClick={onAddRoom}>(+ add room)</button>

      <br/><br/>
      <button onClick={submit}>ADD Home</button>
    </div>
  )
}

