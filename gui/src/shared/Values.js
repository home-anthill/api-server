import React, { useEffect, useState } from 'react';

import { getHeaders } from '../apis/utils';

export default function Values({device}) {
  const [value, setValue] = useState({});

  useEffect(() => {
    async function fn() {
      console.log('useEffect 1 - fn');
      try {
        const response = await fetch(`/api/devices/${device.id}/values`, {
          headers: getHeaders()
        });
        const body = await response.json();
        console.log('Values: ', body);
        setValue(body);
      } catch (err) {
        console.error('Cannot get homes');
      }
    }

    fn();
  }, []);

  async function setOnOff() {
    const response = await fetch(`/api/devices/${device.id}/values/onoff`, {
      method: 'PUT',
      headers: getHeaders(),
      body: JSON.stringify({
        on: value.on
      })
    });
    const body = await response.json();
    console.log('response', body);
  }

  async function setTemperature() {
    const response = await fetch(`/api/devices/${device.id}/values/temperature`, {
      method: 'PUT',
      headers: getHeaders(),
      body: JSON.stringify({
        temperature: value.temperature
      })
    });
    const body = await response.json();
    console.log('response', body);
  }

  async function setMode() {
    const response = await fetch(`/api/devices/${device.id}/values/mode`, {
      method: 'PUT',
      headers: getHeaders(),
      body: JSON.stringify({
        mode: value.mode
      })
    });
    const body = await response.json();
    console.log('response', body);
  }

  async function setFanMode() {
    const response = await fetch(`/api/devices/${device.id}/values/fanmode`, {
      method: 'PUT',
      headers: getHeaders(),
      body: JSON.stringify({
        fanMode: value.fanMode
      })
    });
    const body = await response.json();
    console.log('response', body);
  }

  async function setFanSpeed() {
    const response = await fetch(`/api/devices/${device.id}/values/fanspeed`, {
      method: 'PUT',
      headers: getHeaders(),
      body: JSON.stringify({
        fanSpeed: value.fanSpeed
      })
    });
    const body = await response.json();
    console.log('response', body);
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
