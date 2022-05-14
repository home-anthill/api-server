import React, { useEffect, useState } from 'react';

import { getHeaders } from '../../apis/utils';

import './Values.css';

import { Button, Checkbox, FormControl, FormControlLabel, Switch, TextField } from '@mui/material';

export default function Values({device}) {
  const [onOff, setOnOff] = useState(false);
  const [temperature, setTemperature] = useState(0);
  const [mode, setMode] = useState(0);
  const [fanMode, setFanMode] = useState(0);
  const [fanSpeed, setFanSpeed] = useState(0);

  useEffect(() => {
    async function fn() {
      try {
        const response = await fetch(`/api/devices/${device.id}/values`, {
          headers: getHeaders()
        });
        const body = await response.json();
        console.log('Values: ', body);

        setOnOff(body.on);
        setTemperature(body.temperature);
        setMode(body.mode);
        setFanMode(body.fanMode);
        setFanSpeed(body.fanSpeed);
      } catch (err) {
        console.error('Cannot get homes');
      }
    }

    fn();
  }, []);

  async function postOnOff() {
    const response = await fetch(`/api/devices/${device.id}/values/onoff`, {
      method: 'POST',
      headers: getHeaders(),
      body: JSON.stringify({
        on: onOff
      })
    });
    const body = await response.json();
    console.log('response', body);
  }

  async function postTemperature() {
    const response = await fetch(`/api/devices/${device.id}/values/temperature`, {
      method: 'POST',
      headers: getHeaders(),
      body: JSON.stringify({
        temperature: temperature
      })
    });
    const body = await response.json();
    console.log('response', body);
  }

  async function postMode() {
    const response = await fetch(`/api/devices/${device.id}/values/mode`, {
      method: 'POST',
      headers: getHeaders(),
      body: JSON.stringify({
        mode: mode
      })
    });
    const body = await response.json();
    console.log('response', body);
  }

  async function postFanMode() {
    const response = await fetch(`/api/devices/${device.id}/values/fanmode`, {
      method: 'POST',
      headers: getHeaders(),
      body: JSON.stringify({
        fanMode: fanMode
      })
    });
    const body = await response.json();
    console.log('response', body);
  }

  async function postFanSpeed() {
    const response = await fetch(`/api/devices/${device.id}/values/fanspeed`, {
      method: 'POST',
      headers: getHeaders(),
      body: JSON.stringify({
        fanSpeed: fanSpeed
      })
    });
    const body = await response.json();
    console.log('response', body);
  }

  function handleOnOffChange(event) {
    setOnOff(event.target.checked);
  }

  function handleTemperatureChange(event) {
    setTemperature(+(event.target.value));
  }

  function handleModeChange(event) {
    setMode(+(event.target.value));
  }

  function handleFanModeChange(event) {
    setFanMode(+(event.target.value));
  }

  function handleFanSpeedChange(event) {
    setFanSpeed(+(event.target.value));
  }

  return (
    <div className="Values">

      <div className="ValueContainer">

      </div>

      <div className="ValueContainer">
        <FormControlLabel
          control={
            <Switch
              checked={onOff}
              onChange={e => handleOnOffChange(e)} />
          }
          label="On/Off"
        />
        <Button sx={{
                  marginLeft: '10px'
                }}
                onClick={postOnOff}>Send</Button>
      </div>

      <div className="ValueContainer">
        <FormControl>
          <TextField label="Temperature"
                   variant="outlined"
                   value={temperature}
                   onChange={e => handleTemperatureChange(e)}/>
        </FormControl>
        <Button sx={{
                  marginLeft: '10px'
                }}
                onClick={postTemperature}>Send</Button>
      </div>

      <div className="ValueContainer">
        <FormControl>
          <TextField label="Mode"
                     variant="outlined"
                     value={mode}
                     onChange={e => handleModeChange(e)}/>
        </FormControl>
        <Button sx={{
                  marginLeft: '10px'
                }}
                onClick={postMode}>Send</Button>
      </div>

      <div className="ValueContainer">
        <FormControl>
          <TextField label="FanMode"
                     variant="outlined"
                     value={fanMode}
                     onChange={e => handleFanModeChange(e)}/>
        </FormControl>
        <Button sx={{
                  marginLeft: '10px'
                }}
                onClick={postFanMode}>Send</Button>
      </div>

      <div className="ValueContainer">
        <FormControl>
          <TextField label="FanSpeed"
                     variant="outlined"
                     value={fanSpeed}
                     onChange={e => handleFanSpeedChange(e)}/>
        </FormControl>
        <Button sx={{
                  marginLeft: '10px'
                }}
                onClick={postFanSpeed}>Send</Button>
      </div>
    </div>
  );
}
