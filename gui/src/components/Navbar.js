import {useEffect, useState} from "react";
import { useNavigate } from 'react-router-dom';
import axios from 'axios';

export default function Navbar () {
  const [profile, setProfile] = useState([]);
  const navigate = useNavigate();

  function showProfile() {
    navigate(`/profile`, {state: {profile}});
  }

  useEffect(() => {
    async function fn() {
      let token = localStorage.getItem('token');
      let headers = {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer ' + token
      };

      const response = await axios.get('api/profile', {
        headers
      })
      const data = response.data;
      setProfile(data);
    }

    fn();
  }, []);

  return (
    <nav className="navbar">
      <h1>Navbar title</h1>
      <img src="https://magal.li/i/50/50" alt="profile icon" width="50" height="50" onClick={() => showProfile()}/>
    </nav>
  );
}
