import React  from 'react';
import { useLocation, useNavigate } from 'react-router-dom';

import './HomeDetails.css';

export default function HomeDetails() {
  const { state } = useLocation();
  const navigate = useNavigate();

  const home = state.home;

  function showRoomDetails(home, room) {
    navigate(`/main/homes/${home.id}/rooms/${room.id}`, {state: {home, room}});
  }

  return (
    <div className="App">
      <h1>HomeDetails</h1>
      <div className="home" key={home}>
        <p>{home.name} @ {home.location}</p>
      </div>
      {home.rooms?.map(room => (
        <div className="room" key={room} onClick={() => showRoomDetails(home, room)}>
          <p>{room.name} @ {room.floor}</p>
        </div>
      ))}
    </div>
  )
}

