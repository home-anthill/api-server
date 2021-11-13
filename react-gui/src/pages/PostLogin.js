import { useNavigate, useLocation } from 'react-router-dom';
import { useEffect } from 'react';
import useAuth from '../useAuth';

export default function PostLogin() {
  const location = useLocation();
  const navigate = useNavigate();
  let auth = useAuth();

  useEffect(() => {
    const token = new URLSearchParams(location.search).get('token');
    auth.login(token);
    navigate('/main');
  });

  return (
    <div>
      <h1>PostLogin</h1>
      <div>{location.pathname}</div>
    </div>
  )
}
