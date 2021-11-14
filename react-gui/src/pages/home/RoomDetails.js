import React from 'react';
import { useLocation } from 'react-router-dom';

export default function RoomDetails() {
  const { state } = useLocation();
  const home = state.home;
  const room = state.room;

  return (
    <div className="App">
      <h1>RoomDetails from home {home.name}</h1>
      <div className="room" key={room}>
        <p>{room.name} @ {room.floor}</p>
      </div>
    </div>
  )
}
