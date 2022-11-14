import React from 'react';

import { Card, CardActions, CardContent, Typography, IconButton } from '@mui/material';
import SettingsIcon from '@mui/icons-material/Settings';
import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import AutoStoriesIcon from '@mui/icons-material/AutoStories';

import './Device.css';

export default function Device({device, deviceType, onShowSettings, onShowController, onShowSensor}) {

  return (
    <Card variant="outlined"
          key={device?.id}
          sx={{margin: '12px', minWidth: '250px'}}>
      <CardContent sx={{paddingTop: '14px', paddingLeft: '14px', paddingRight: '14px', paddingBottom: '6px'}}>
        <Typography sx={{fontSize: 14}} component="div">
          {device?.mac}
        </Typography>
        <Typography sx={{fontSize: 12}} color="text.secondary" gutterBottom>
          {device?.model} - {device?.manufacturer}
        </Typography>
        <div className="DeviceDivider"></div>
      </CardContent>
      <CardActions>
        <IconButton aria-label="settings" onClick={() => onShowSettings(device)}>
          <SettingsIcon/>
        </IconButton>
        {deviceType === 'controller' &&
          <IconButton aria-label="controller" onClick={() => onShowController(device)}>
            <PlayArrowIcon/>
          </IconButton>
        }
        {deviceType === 'sensor' &&
          <IconButton aria-label="sensor" onClick={() => onShowSensor(device)}>
            <AutoStoriesIcon/>
          </IconButton>
        }
      </CardActions>
    </Card>
  )
}

