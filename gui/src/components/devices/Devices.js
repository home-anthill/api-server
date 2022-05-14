import React  from 'react';
import { useNavigate } from 'react-router-dom';

import { Typography, IconButton, CardContent, CardActions, Card } from '@mui/material';
import SettingsIcon from '@mui/icons-material/Settings';

import useDevices from '../../apis/useDevices';

import './Devices.css';

export default function Devices() {
  const {
    data: devicesData,
    loading: devicesLoading,
    error: devicesError,
  } = useDevices();

  const navigate = useNavigate();

  function showDeviceDetails(device) {
    navigate(`/main/devices/${device.id}`, {state: {device}});
  }

  return (
    <div className="Devices">
      <Typography variant="h2" component="h1">
        Devices
      </Typography>
      <div className="HomesContainer">
      {devicesError ? (
        <div className="error">
          Something went wrong:
          <div className="error-contents">
            {devicesError.message}
          </div>
        </div>
      ) : devicesLoading ? (
        <div className="loading">Loading...</div>
      ) : devicesData && devicesData.length ? (
        <>
          {devicesData.map((device) => (
            <Card variant="outlined"
                  key={device.id}
                  sx={{
                    margin: "12px",
                    minWidth: "250px"
                  }}>
              <CardContent>
                <Typography variant="h5" component="div">
                  {device.name}
                </Typography>
                <Typography sx={{ fontSize: 14 }} color="text.secondary" gutterBottom>
                  {device.manufacturer}
                </Typography>
                <Typography sx={{ fontSize: 14 }} color="text.secondary" gutterBottom>
                  {device.model}
                </Typography>
              </CardContent>
              <CardActions>
                <IconButton aria-label="settings" onClick={() => showDeviceDetails(device)}>
                  <SettingsIcon />
                </IconButton>
              </CardActions>
            </Card>
          ))}
        </>
      ) : (
        'No devices'
      )}
      </div>
    </div>
  )
}

