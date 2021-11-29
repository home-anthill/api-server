import React from "react";
import axios from 'axios';

export default function Values ({device}) {
  let token = localStorage.getItem('token');
  let headers = {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer ' + token
  };

  async function setOnOff() {
    const response = await axios.post(`http://localhost:8082/api/devices/${device.id}/values/onoff`, {
      on: true
    }, {
      headers
    })
    const data = response.data;
    console.log('response', data);
  }

  async function setTemperature() {
    const response = await axios.post(`http://localhost:8082/api/devices/${device.id}/values/temperature`, {
      temperature: 25
    }, {
      headers
    })
    const data = response.data;
    console.log('response', data);
  }

  async function setMode() {
    const response = await axios.post(`http://localhost:8082/api/devices/${device.id}/values/mode`, {
      mode: 1
    }, {
      headers
    })
    const data = response.data;
    console.log('response', data);
  }

  async function setFanMode() {
    const response = await axios.post(`http://localhost:8082/api/devices/${device.id}/values/fanmode`, {
      fan: 2
    }, {
      headers
    })
    const data = response.data;
    console.log('response', data);
  }

  async function setFanSwing() {
    const response = await axios.post(`http://localhost:8082/api/devices/${device.id}/values/fanswing`, {
      swing: true
    }, {
      headers
    })
    const data = response.data;
    console.log('response', data);
  }

  // TODO add current values (probably with a get via REST to api-server and than to api-devices via gRPC to get device[i].status
  return (
    <div>
      <h1>Device</h1>
      <p>{device.name} - {device.manufacturer} - {device.model}</p>
      <br/>
      <button onClick={setOnOff}>Set On/Off</button>
      <button onClick={setTemperature}>Set Temperature</button>
      <button onClick={setMode}>Set Mode</button>
      <button onClick={setFanMode}>Set Fan mode</button>
      <button onClick={setFanSwing}>Set Fan swing</button>
    </div>
  );
}
