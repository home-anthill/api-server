import React, { useEffect, useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import axios from 'axios';

export default function EditHome() {
  const { state } = useLocation();
  const homeInput = state.home;

  const [home, setHome] = useState({});
  const [rooms, setRooms] = useState([]);

  const navigate = useNavigate();

  useEffect(() => {
    console.log('homeInput - setting states. ', homeInput);
    setHome(homeInput);
    setRooms(homeInput.rooms);
  }, [homeInput]);

  const submit = async e => {
    e.preventDefault();

    let token = localStorage.getItem('token');
    let headers = {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer ' + token
    };
    try {
      const response = await axios.put(`http://localhost:8082/api/homes/${home.id}`, {
        name: home.name,
        location: home.location,
        rooms: rooms
      }, {
        headers
      });
      console.log('response', response);
      // navigate back
      navigate(-1);
    } catch (err) {
      console.error('Cannot add a new home');
    }
  }

  function onChangeRoom(index, newRoom) {
    const updated = [...rooms];
    updated[index] = newRoom;
    setRooms(updated)
  }

  function onAddRoom() {
    console.log('add room');
    setRooms([...rooms, {name: '', floor: 0}]);
  }

  function onRemoveRoom(index) {
    let updated = [...rooms];
    updated = updated.filter((room, i) => i !== index);
    setRooms(updated)
  }

  return (
    <div className="App">
      <h1>Edit Home</h1>
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
          <button onClick={() => onRemoveRoom(index)}>X</button>
        </>
      ))}
      <button onClick={onAddRoom}>(+ add room)</button>

      <br/><br/>
      <button onClick={submit}>Save Home</button>
    </div>
  )
}

