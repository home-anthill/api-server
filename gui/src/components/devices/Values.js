import React, { useEffect, useState } from 'react';

import { Button, FormControl, FormControlLabel, Switch, TextField } from '@mui/material';

import './Values.css';

import { getApi, postApi } from '../../apis/api';

export default function Values({device}) {
  const [onOff, setOnOff] = useState(false);
  const [temperature, setTemperature] = useState(0);
  const [mode, setMode] = useState(0);
  const [fanMode, setFanMode] = useState(0);
  const [fanSpeed, setFanSpeed] = useState(0);

  useEffect(() => {
    async function fn() {
      try {
        const response = await getApi(`/api/devices/${device.id}/values`);
        setOnOff(response.on);
        setTemperature(response.temperature);
        setMode(response.mode);
        setFanMode(response.fanMode);
        setFanSpeed(response.fanSpeed);
      } catch (err) {
        console.error('Cannot get homes');
      }
    }

    fn();
  }, []);

  async function postOnOff() {
    await postApi(`/api/devices/${device.id}/values/onoff`, {
      on: onOff
    });
  }

  async function postTemperature() {
    await postApi(`/api/devices/${device.id}/values/temperature`, {
      temperature: temperature
    });
  }

  async function postMode() {
    await postApi(`/api/devices/${device.id}/values/mode`, {
      mode: mode
    });
  }

  async function postFanMode() {
    await postApi(`/api/devices/${device.id}/values/fanmode`, {
      fanMode: fanMode
    });
  }

  async function postFanSpeed() {
    await postApi(`/api/devices/${device.id}/values/fanspeed`, {
      fanSpeed: fanSpeed
    });
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
