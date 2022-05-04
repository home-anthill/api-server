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
      const response = await axios.put(`/api/homes/${home.id}`, {
        name: home.name,
        location: home.location
        // cannot change room with this api
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
    updated[index].name = newRoom.name;
    updated[index].floor = newRoom.floor;
    setRooms(updated)
  }

  function onAddRoom() {
    console.log('add room');
    setRooms([...rooms, {name: '', floor: 0}]);
  }

  async function onRemoveRoom(room) {
    let token = localStorage.getItem('token');
    let headers = {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer ' + token
    };
    try {
      const response = await axios.delete(`/api/homes/${home.id}/rooms/${room.id}`, {
        headers
      });
      console.log('response', response);
      // navigate back
      navigate(-1);
    } catch (err) {
      console.error('Cannot add a new home');
    }
  }

  async function onSaveRoom(room) {
    let token = localStorage.getItem('token');
    let headers = {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer ' + token
    };
    try {
      let response;
      if (room.id) {
        response = await axios.put(`/api/homes/${home.id}/rooms/${room.id}`, {
          'name': room.name,
          'floor': room.floor
        }, {
          headers
        });
      } else {
        response = await axios.post(`/api/homes/${home.id}/rooms`, {
          'name': room.name,
          'floor': room.floor
        }, {
          headers
        });
      }
      console.log('response', response);
      // navigate back
      navigate(-1);
    } catch (err) {
      console.error('Cannot add a new home');
    }
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
      <br/>
      <button onClick={submit}>Save Home</button>
      <br/>

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
          <button onClick={() => onSaveRoom(room)}>Save</button>
          <button onClick={() => onRemoveRoom(room)}>Delete</button>
        </>
      ))}
      <br/>
      <br/>
      <button onClick={onAddRoom}>(+ add room)</button>

      <br/><br/>
    </div>
  )
}

