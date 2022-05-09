import React, { useState } from 'react';
import { useLocation } from 'react-router-dom';

import { getHeaders } from '../apis/utils';
import { Avatar, Button, Typography } from '@mui/material';

import './Profile.css';

export default function Profile() {
  const {state} = useLocation();
  const profile = state.profile;

  const [apiToken, setApiToken] = useState('********-****-****-****-************');

  async function regenerateApiToken() {
    try {
      const response = await fetch(`/api/profiles/${profile.id}/tokens`, {
        method: 'POST',
        headers: getHeaders(),
        body: JSON.stringify({})
      });
      const body = await response.json();
      console.log('ApiToken response: ', body);
      setApiToken(body.apiToken);
    } catch (err) {
      console.error('Cannot re-generate API Token');
    }
  }

  return (
    <div className="Profile">
      <Typography variant="h2" component="h1">
        Profile
      </Typography>
      <div className="ProfileContainer">
        <Typography variant="h5" component="div" gutterBottom>
          {profile.github?.login}
        </Typography>
        <Typography variant="h5" component="div" gutterBottom>
          {profile.github?.name}
        </Typography>
        <Typography sx={{ fontSize: 12 }} variant="h5" component="div" gutterBottom>
          {profile.github?.email}
        </Typography>
        <br />
        <Avatar
          alt="profile"
          src={profile.github?.avatarURL}
          sx={{ width: 256, height: 256 }}
        />
        <br />
        <p>{apiToken}</p>
        <Button onClick={regenerateApiToken}>Regenerate APIToken</Button>
      </div>
    </div>
  )
}
