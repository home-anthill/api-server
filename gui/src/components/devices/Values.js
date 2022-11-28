import React, { useEffect, useState } from 'react';

import { Button, Snackbar, FormControl, FormControlLabel, Switch, InputLabel, Select, MenuItem } from '@mui/material';
import MuiAlert from '@mui/material/Alert';

import './Values.css';

import { getApi, postApi } from '../../apis/api';

const STATE_SUCCESS = 'Device state update successfully!';
const STATE_ERROR = 'Cannot update device state!';

export default function Values({device}) {
  const [onOff, setOnOff] = useState(false);
  const [temperature, setTemperature] = useState(28);
  const [mode, setMode] = useState(1);
  const [fanSpeed, setFanSpeed] = useState(1);
  const [snackBarState, setSnackBarState] = useState({
    open: false,
    severity: 'success',
    message: STATE_SUCCESS
  });

  // use a custom Alert extending MuiAlert instead of the Alert defined in @mui/material
  const Alert = React.forwardRef(function Alert(props, ref) {
    return <MuiAlert elevation={6} ref={ref} variant="filled" {...props} />;
  });

  useEffect(() => {
    async function fn() {
      try {
        const response = await getApi(`/api/devices/${device.id}/values`);
        setOnOff(response.on);
        setTemperature(response.temperature);
        setMode(response.mode);
        setFanSpeed(response.fanSpeed);
      } catch (err) {
        console.error('Cannot get device values');
      }
    }

    fn();
  }, [device.id]);

  async function postValues() {
    try {
      await postApi(`/api/devices/${device.id}/values`, {
        on: onOff,
        temperature: temperature,
        mode: mode,
        fanSpeed: fanSpeed
      });
      setSnackBarState({...snackBarState, open: true, severity: 'success', message: STATE_SUCCESS});
    } catch (err) {
      console.log('Cannot set device values. Err = ', err);
      setSnackBarState({...snackBarState, open: true, severity: 'error', message: STATE_ERROR});
    }
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

  function handleFanSpeedChange(event) {
    setFanSpeed(+(event.target.value));
  }

  const handleSnackbarClose = (event, reason) => {
    if (reason === 'clickaway') {
      return;
    }
    setSnackBarState({...snackBarState, open: false});
  };

  return (
    <div className="Values">
      <div className="ValueContainer">
        <FormControlLabel
          control={
            <Switch
              checked={onOff}
              onChange={e => handleOnOffChange(e)}/>
          }
          label="On/Off"
        />
      </div>
      <div className="ValueContainer">
        <FormControl fullWidth>
          <InputLabel id="temperature-select-label">Temperature</InputLabel>
          <Select
            labelId="temperature-select-label"
            id="temperature-select"
            value={temperature}
            label="Temperature"
            onChange={e => handleTemperatureChange(e)}
          >
            {[17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30].map(item => (
              <MenuItem value={item}>{item}</MenuItem>
            ))}
          </Select>
        </FormControl>
      </div>
      <div className="ValueContainer">
        <FormControl fullWidth>
          <InputLabel id="mode-select-label">Mode</InputLabel>
          <Select
            labelId="mode-select-label"
            id="mode-select"
            value={mode}
            label="Mode"
            onChange={e => handleModeChange(e)}
          >
            <MenuItem value={1}>Cool</MenuItem>
            <MenuItem value={2}>Auto</MenuItem>
            <MenuItem value={3}>Heat</MenuItem>
            <MenuItem value={4}>Fan</MenuItem>
            <MenuItem value={5}>Dry</MenuItem>
          </Select>
        </FormControl>
      </div>
      <div className="ValueContainer">
        <FormControl fullWidth>
          <InputLabel id="fan-select-label">Fan</InputLabel>
          <Select
            labelId="fan-select-label"
            id="fan-select"
            value={fanSpeed}
            label="Fan"
            onChange={e => handleFanSpeedChange(e)}
          >
            <MenuItem value={1}>Min</MenuItem>
            <MenuItem value={2}>Med</MenuItem>
            <MenuItem value={3}>Max</MenuItem>
            <MenuItem value={4}>Auto</MenuItem>
            <MenuItem value={5}>Auto0</MenuItem>
          </Select>
        </FormControl>
      </div>

      <Button variant="contained"
              color="success"
              sx={{marginTop: '10px'}}
              onClick={postValues}>Send</Button>

      <Snackbar open={snackBarState.open}
                autoHideDuration={2000}
                onClose={handleSnackbarClose}
                anchorOrigin={{vertical: 'bottom', horizontal: 'center'}}>
        <Alert onClose={handleSnackbarClose} severity={snackBarState.severity} sx={{width: '100%'}}>
          {snackBarState.message}
        </Alert>
      </Snackbar>
    </div>
  );
}
