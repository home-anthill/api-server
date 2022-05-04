import { useEffect, useState } from 'react'

function getHeaders() {
  let token = localStorage.getItem('token');
  let headers = {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer ' + token
  };
  return headers;
}

const useDevices = () => {
  const [data, setData] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState();

  useEffect(() => {
    let didCancel = false;
    setError(null);
    (async () => {
      try {
        setLoading(true);
        const response = await fetch('/api/devices', {
          headers: getHeaders()
        })
        if (!response.ok) {
          const text = await response.text();
          throw new Error(`Unable to read devices: ${text}`);
        }
        const body = await response.json();
        if (!didCancel) {
          setData(body);
        }
      } catch (err) {
        setError(err);
      } finally {
        setLoading(false);
      }
    })()
    return () => {
      didCancel = true;
    }
  }, [])
  return { data, loading, error }
}

export default useDevices
