import React, { useEffect, useState } from 'react';

import { Button, Snackbar, FormControl, FormControlLabel, Switch, TextField } from '@mui/material';
import MuiAlert from '@mui/material/Alert';

import './Values.css';

import { getApi, postApi } from '../../apis/api';

const STATE_SUCCESS = 'Device state update successfully!';
const STATE_ERROR = 'Cannot update device state!';

export default function Values({device}) {
  const [onOff, setOnOff] = useState(false);
  const [temperature, setTemperature] = useState(0);
  const [mode, setMode] = useState(0);
  const [fanSpeed, setFanSpeed] = useState(0);
  const [swing, setSwing] = useState(false);
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
        setSwing(response.swing);
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
        fanSpeed: fanSpeed,
        swing: swing,
      });
      setSnackBarState({...snackBarState, open: true, severity: 'success', message: STATE_SUCCESS});
    } catch(err) {
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

  function handleSwingChange(event) {
    setSwing(event.target.checked);
  }

  const handleSnackbarClose = (event, reason) => {
    if (reason === 'clickaway') {
      return;
    }
    setSnackBarState({...snackBarState, open: false });
  };

  return (
    <div className="Values">
      <div className="ValueContainer">
        <FormControlLabel
          control={
            <Switch
              checked={onOff}
              onChange={e => handleOnOffChange(e)} />
          }
          label="On/Off"
        />
      </div>
      <div className="ValueContainer">
        <FormControl>
          <TextField label="Temperature"
                   variant="outlined"
                   value={temperature}
                   onChange={e => handleTemperatureChange(e)}/>
        </FormControl>
      </div>
      <div className="ValueContainer">
        <FormControl>
          <TextField label="Mode"
                     variant="outlined"
                     value={mode}
                     onChange={e => handleModeChange(e)}/>
        </FormControl>
      </div>
      <div className="ValueContainer">
        <FormControl>
          <TextField label="FanSpeed"
                     variant="outlined"
                     value={fanSpeed}
                     onChange={e => handleFanSpeedChange(e)}/>
        </FormControl>
      </div>
      <div className="ValueContainer">
        <FormControlLabel
          control={
            <Switch
              checked={swing}
              onChange={e => handleSwingChange(e)} />
          }
          label="Swing"
        />
      </div>

      <Button variant="contained"
              color="success"
              sx={{ marginTop: '10px'}}
              onClick={postValues}>Send</Button>

      <Snackbar open={snackBarState.open}
                autoHideDuration={2000}
                onClose={handleSnackbarClose}
                anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}>
        <Alert onClose={handleSnackbarClose} severity={snackBarState.severity} sx={{ width: '100%' }}>
          {snackBarState.message}
        </Alert>
      </Snackbar>
    </div>
  );
}
