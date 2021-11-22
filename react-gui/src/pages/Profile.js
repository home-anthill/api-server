import React from 'react';
import { useLocation } from 'react-router-dom';

export default function Profile() {
  const {state} = useLocation();
  const profile = state.profile?.profile;

  return (
    <div className="App">
      <h1>Profile</h1>
      <p>Login: {profile.github?.login}</p>
      <p>Name: {profile.github?.name}</p>
      <p>Email: {profile.github?.email}</p>
      <img src={profile.github?.avatarURL} />
    </div>
  )
}

