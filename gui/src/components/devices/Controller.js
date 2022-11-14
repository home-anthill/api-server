import React from 'react';
import { useLocation } from 'react-router-dom';

import { Typography } from '@mui/material';

import './Controller.css';

import Values from './Values';

export default function Controller() {
  const {state} = useLocation();
  const device = state.device;

  return (
    <div className="Controller">
      <Typography variant="h2" component="h1">
        Controller
      </Typography>
      <div className="ControllerContainer">
        <Typography variant="h5" component="h2">
          {device?.mac}
        </Typography>
        <Typography variant="subtitle1" component="h3">
          {device?.manufacturer} - {device?.model}
        </Typography>

        <Values device={device}/>
      </div>
    </div>
  )
}

