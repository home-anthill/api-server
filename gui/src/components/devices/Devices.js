import React, { useNavigate } from 'react-router-dom';

import './Devices.css';

import useDevices from '../../apis/useDevices';

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
    <div className="App">
      <h1>Devices</h1>
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
            <div className="device" key={device.id}>
              <p onClick={() => showDeviceDetails(device)}>{device.name} - {device.manufacturer} - {device.model}</p>
            </div>
          ))}
        </>
      ) : (
        'No devices'
      )}
    </div>
  )
}

