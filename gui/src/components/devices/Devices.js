import React  from 'react';
import { useNavigate } from 'react-router-dom';

import { Typography } from '@mui/material';

import useDevices from '../../apis/useDevices';
import Device from '../../shared/Device';

import './Devices.css';

export default function Devices() {
  const navigate = useNavigate();
  // get devices in a object where devices are grouped by homes and rooms and
  // are separated by their types (controllers or sensors) in an array called `homeDevices`.
  // Devices that are not assigned are defined in `unassignedDevices` array
  const {
    data: devicesResult,
    loading: devicesLoading,
    error: devicesError,
  } = useDevices();

  function showDeviceSettings(device) {
    if (!device) {
      console.error(`Cannot show settings - 'id' is missing`);
      return;
    }
    navigate(`/main/devices/${device.id}`, {state: {device}});
  }

  function showController(device) {
    if (!device) {
      console.error(`Cannot open controller - 'id' is missing`);
      return;
    }
    navigate(`/main/devices/${device.id}/controller`, {state: {device}});
  }

  function showSensor(device) {
    if (!device) {
      console.error(`Cannot open sensor - 'id' is missing`);
      return;
    }
    navigate(`/main/devices/${device.id}/sensor`, {state: {device}});
  }

  function isController (device) {
    return device.features.find(feature => feature.type === 'controller') !== undefined;
  }

  return (
    <div className="DevicesContainer">
      <Typography variant="h2" component="h1" textAlign={'center'}>
        Devices
      </Typography>
      {devicesError ? (
        <div className="error">
          Something went wrong:
          <div className="error-contents">
            {devicesError.message}
          </div>
        </div>
      ) : devicesLoading ? (
        <div className="loading">Loading...</div>
      ) : devicesResult?.homeDevices?.length ? (
        <>
          {devicesResult.unassignedDevices &&
            <>
              <div className="HomeContainer">
                <Typography variant="h5" component="h1">
                  Unassigned
                </Typography>
                <div className="FeaturesContainer">
                  {devicesResult.unassignedDevices.map((device) => (
                    <Device device={device}
                            deviceType={isController(device) ? 'controller' : 'sensor'}
                            onShowController={() => showController(device)}
                            onShowSensor={() => showSensor(device)}
                            onShowSettings={() => showDeviceSettings(device)}></Device>
                  ))}
                </div>
              </div>
              <div className="DevicesDivider"></div>
            </>
          }

          {devicesResult.homeDevices.map((home) => (
            <>
              <div className="HomeContainer">
                <Typography variant="h5" component="h1">
                  { home.name } ({ home.location })
                </Typography>
                <br />
                {home.rooms.map((room) => (
                  <div className="RoomContainer">
                    <Typography variant="h6" component="h2">
                      { room.name } - { room.floor }
                    </Typography>
                    {(room.controllerDevices.length > 0 || room.sensorDevices.length > 0) ? (
                      <>
                        <div className="FeaturesContainer">
                          {room.controllerDevices.map((controller) => (
                            <Device device={controller}
                                    deviceType={'controller'}
                                    onShowController={() => showController(controller)}
                                    onShowSettings={() => showDeviceSettings(controller)}></Device>
                          ))}
                        </div>
                        <div className="FeaturesContainer">
                          {room.sensorDevices.map((sensor) => (
                            <Device device={sensor}
                                    deviceType={'sensor'}
                                      onShowSensor={() => showSensor(sensor)}
                                    onShowSettings={() => showDeviceSettings(sensor)}></Device>
                          ))}
                        </div>
                      </>
                    ) : (
                      'No devices to show'
                    )}
                  </div>
                ))}
              </div>
              <div className="DevicesDivider"></div>
            </>
          ))}
        </>
      ) : (
        'No data to show'
      )}
    </div>
  )
}

