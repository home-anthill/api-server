import React from 'react';
import { useLocation } from 'react-router-dom';

export default function Profile() {
  const {state} = useLocation();

  const profile = state.profile?.user;

  return (
    <div className="App">
      <h1>Profile</h1>
      <p>{profile.login}</p>
      <p>{profile.name}</p>
    </div>
  )
}

