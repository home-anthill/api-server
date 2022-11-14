import React from 'react';
import { useLocation } from 'react-router-dom';

import { Typography } from '@mui/material';

import './Sensor.css';

import useSensorValues from '../../apis/useSensorValues';
import SensorValue from '../../shared/SensorValue';

export default function Sensor() {
  const {state} = useLocation();
  const sensor = state.device;

  const {
    data: sensorValuesResult,
    loading: sensorValuesLoading,
    error: sensorValuesError,
  } = useSensorValues(sensor);

  return (
    <div className="SensorContainer">
      <Typography variant="h2" component="h1">
        Sensor
      </Typography>
      <div className="Sensor">
        <Typography variant="h5" component="h2">
          {sensorValuesResult?.mac}
        </Typography>
        <Typography variant="subtitle1" component="h3">
          {sensorValuesResult?.manufacturer} - {sensorValuesResult?.model}
        </Typography>
        {sensorValuesError ? (
          <div className="error">
            Something went wrong:
            <div className="error-contents">
              {sensorValuesError.message}
            </div>
          </div>
        ) : sensorValuesLoading ? (
          <div className="loading">Loading...</div>
        ) : sensorValuesResult !== undefined ? (
          <div className="SensorFeaturesContainer">
            {sensorValuesResult?.features?.map(feature => (
              <div className="FeatureValue" key={feature?.uuid}>
                <SensorValue sensorFeatureValue={feature}></SensorValue>
              </div>
            ))}
          </div>
        ) : (
          'No data to show'
        )}
      </div>
    </div>
  )
}

