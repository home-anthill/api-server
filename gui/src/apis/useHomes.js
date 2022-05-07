import { useEffect, useState } from 'react'
import { getHeaders } from './utils';

const useHomes = (homes) => {
  const [data, setData] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState();

  useEffect(() => {
    console.log('trigegred homes:', homes);
    let didCancel = false;
    setError(null);
    (async () => {
      try {
        setLoading(true);
        const response = await fetch('/api/homes', {
          headers: getHeaders()
        })
        if (!response.ok) {
          const text = await response.text();
          throw new Error(`Unable to read homes: ${text}`);
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
    })();
    return () => {
      didCancel = true;
    }
  }, [homes])
  return {data, loading, error}
}

export default useHomes
