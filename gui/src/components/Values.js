import React, { useEffect, useState } from 'react';
import axios from 'axios';

export default function Values({device}) {
  let token = localStorage.getItem('token');
  let headers = {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer ' + token
  };

  const [value, setValue] = useState({});

  useEffect(() => {
    async function fn() {
      console.log('useEffect 1 - fn');
      const token = localStorage.getItem('token');
      const headers = {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer ' + token
      };
      try {
        const response = await axios.get(`api/devices/${device.id}/values`, {
          headers
        })
        const data = response.data;
        console.log('Values: ', data);
        setValue(data);
      } catch (err) {
        console.error('Cannot get homes');
      }
    }

    fn();
  }, []);

  async function setOnOff() {
    const response = await axios.post(`api/devices/${device.id}/values/onoff`, {
      on: value.on
    }, {
      headers
    })
    const data = response.data;
    console.log('response', data);
  }

  async function setTemperature() {
    const response = await axios.post(`api/devices/${device.id}/values/temperature`, {
      temperature: value.temperature
    }, {
      headers
    })
    const data = response.data;
    console.log('response', data);
  }

  async function setMode() {
    const response = await axios.post(`api/devices/${device.id}/values/mode`, {
      mode: value.mode
    }, {
      headers
    })
    const data = response.data;
    console.log('response', data);
  }

  async function setFanMode() {
    const response = await axios.post(`api/devices/${device.id}/values/fanmode`, {
      fanMode: value.fanMode
    }, {
      headers
    })
    const data = response.data;
    console.log('response', data);
  }

  async function setFanSpeed() {
    const response = await axios.post(`api/devices/${device.id}/values/fanspeed`, {
      fanSpeed: value.fanSpeed
    }, {
      headers
    })
    const data = response.data;
    console.log('response', data);
  }

  return (
    <>
      <div>
        <label>
          <input
            type="checkbox"
            checked={value.on}
            onChange={event => setValue(Object.assign({}, value, {on: event.target.checked}))}/>
          On
        </label>
        <br/>
        <button onClick={setOnOff}>Set On/Off</button>
      </div>
      <div>
        <input
          value={value.temperature}
          onChange={event => setValue(Object.assign({}, value, {temperature: +(event.target.value)}))}
          type="number"
          placeholder="Temperature"
        />
        <br/>
        <button onClick={setTemperature}>Set Temperature</button>
      </div>
      <div>
        <input
          value={value.mode}
          onChange={event => setValue(Object.assign({}, value, {mode: +(event.target.value)}))}
          type="number"
          placeholder="Mode"
        />
        <br/>
        <button onClick={setMode}>Set Mode</button>
      </div>
      <div>
        <input
          value={value.fanMode}
          onChange={event => setValue(Object.assign({}, value, {fanMode: +(event.target.value)}))}
          type="number"
          placeholder="Fan mode"
        />
        <br/>
        <button onClick={setFanMode}>Set Fan mode</button>
      </div>
      <div>
        <input
          value={value.fanSpeed}
          onChange={event => setValue(Object.assign({}, value, {fanSpeed: +(event.target.value)}))}
          type="number"
          placeholder="Fan speed"
        />
        <br/>
        <button onClick={setFanSpeed}>Set Fan speed</button>
      </div>
    </>
  );
}
