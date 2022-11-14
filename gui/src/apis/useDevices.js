import { useEffect, useState } from 'react'

import { getApi } from './api';

const useDevices = () => {
  const [data, setData] = useState({});
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState();

  useEffect(() => {
    let didCancel = false;
    setError(null);
    (async () => {
      try {
        setLoading(true);
        // get homes and devices to build the model
        const homes = await getApi('/api/homes');
        const devices = await getApi('/api/devices');
        if (!didCancel) {
          const result = {
            unassignedDevices: [],
            homeDevices: []
          }
          // 1) add unassigned devices to `result.unassignedDevices`
          result.unassignedDevices = getUnassignedDevices(homes, devices);
          // 2) add assigned devices with homes and rooms to `result.homeDevices`
          homes.forEach(home => {
            const homeObj = Object.assign({}, home);
            const roomsObjs = [];
            homeObj.rooms.forEach(room => {
              // if this room has devices, otherwise skip it
              if (room && room.devices && room.devices.length > 0) {
                const roomObj = Object.assign({}, room);
                // get all devices in this room removing duplicates
                // and mapping these as device object instead of an id string
                let roomDevices = roomObj.devices
                  // remove duplicated
                  .filter((v1, i, array) => array.findIndex(v2 => (v2 === v1)) === i)
                  // map device id to its full device object
                  .map(deviceId => devices.find(device => device && device.id === deviceId));

                // split those devices into 2 different arrays:
                // - controllers (devices able to receive commands)
                // - sensors (read-only devices)
                roomObj.controllerDevices = getControllers(roomDevices);
                roomObj.sensorDevices = getSensors(roomDevices);
                // remove the list of device ids in string format, because above we added full objects
                delete roomObj.devices;
                // add this room to the list of rooms of the current home
                roomsObjs.push(roomObj);
              }
            });
            // if this home has rooms (added in the loop above), otherwise skip it
            if (roomsObjs.length > 0) {
              homeObj.rooms = roomsObjs;
              result.homeDevices.push(homeObj);
            }
          });
          // 3) result object contains `unassignedDevices`,
          //    `homeDevices` (`controllerDevices` and `sensorDevices`)
          setData(result);
        }
      } catch (err) {
        console.error(err);
        setError(err);
      } finally {
        setLoading(false);
      }
    })()
    return () => {
      didCancel = true;
    }
  }, [])
  return {data, loading, error}
}

function getUnassignedDevices(homes, devices) {
  const rooms = homes
    .filter(home => home.rooms && home.rooms.length > 0)
    .map(home => home.rooms)
    .flat();
  const devicesIds = rooms
    .filter(room => room.devices && room.devices.length > 0)
    .map(room => room.devices)
    .flat();
  return devices
    .filter(device => !devicesIds.includes(device.id));
}

function getControllers(devices) {
  // if a device has a controller feature, it's a controller and it cannot have any sensor feature!
  return devices.filter(device => device.features.find(feature => feature.type === 'controller'));
}

function getSensors(devices) {
  // is a device has only sensor feature, it's a sensor
  return devices.filter(device => device.features.every(feature => feature.type === 'sensor'));
}

export default useDevices
