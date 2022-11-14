import { useEffect, useState } from 'react'

import { getApi } from './api';

const useSensorValues = (sensor) => {
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
        const sensorValues = await getApi(`/api/devices/${sensor.id}/values`);
        if (!didCancel) {
          // TODO create a better object to show more info on the UI
          let sensorObj = Object.assign({}, sensor);
          sensorObj.features = sensorObj
            .features
            // order by priority
            .sort((a, b) => a.priority - b.priority)
            .map(feature => {
              const feat = Object.assign({}, feature);
              const val = sensorValues.find(sensorValue => sensorValue.uuid === feature.uuid);
              feat.value = val.value;
              return feat;
            });
          setData(sensorObj);
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
  }, [sensor])
  return {data, loading, error}
}

export default useSensorValues
