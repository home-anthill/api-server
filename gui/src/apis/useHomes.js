import { useEffect, useState } from 'react'

import { getApi } from './api';

const useHomes = (homes) => {
  const [data, setData] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState();

  useEffect(() => {
    let didCancel = false;
    setError(null);
    (async () => {
      try {
        setLoading(true);
        const response = await getApi('/api/homes');
        if (!didCancel) {
          setData(response);
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
