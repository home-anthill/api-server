import React, { useState } from 'react';
import { useLocation } from 'react-router-dom';
import axios from 'axios';

export default function Profile() {
  const {state} = useLocation();
  const profile = state.profile?.profile;

  const [apiToken, setApiToken] = useState('********-****-****-****-************');

  async function regenerateApiToken() {
    try {
      const token = localStorage.getItem('token');
      const headers = {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer ' + token
      };
      const response = await axios.post(`api/profiles/${profile.id}/tokens`,
        {},
        {
          headers
        }
      );
      const data = response.data;
      console.log('ApiToken response: ', data);
      setApiToken(data.apiToken);
    } catch (err) {
      console.error('Cannot re-generate API Token');
    }
  }

  return (
    <div className="App">
      <h1>Profile</h1>
      <p>Login: {profile.github?.login}</p>
      <p>Name: {profile.github?.name}</p>
      <p>Email: {profile.github?.email}</p>
      <img src={profile.github?.avatarURL} />
      <br />
      <p>{apiToken}</p>
      <button onClick={regenerateApiToken}>Regenerate APIToken</button>
    </div>
  )
}

