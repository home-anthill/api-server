import React, {useEffect, useState} from "react";
import { useNavigate } from 'react-router-dom';

import './Navbar.css';
import logoPng from '../air-conditioner.png'

import AppBar from '@mui/material/AppBar';
import Box from '@mui/material/Box';
import Toolbar from '@mui/material/Toolbar';
import IconButton from '@mui/material/IconButton';
import Typography from '@mui/material/Typography';
import Menu from '@mui/material/Menu';
import MenuIcon from '@mui/icons-material/Menu';
import Container from '@mui/material/Container';
import Avatar from '@mui/material/Avatar';
import Button from '@mui/material/Button';
import MenuItem from '@mui/material/MenuItem';
import { getHeaders } from '../apis/utils';

export default function Navbar () {
  const [profile, setProfile] = useState([]);
  const [anchorElNav, setAnchorElNav] = useState(null);
  const navigate = useNavigate();

  function showProfile() {
    navigate(`/profile`, {state: {profile}});
  }

  useEffect(() => {
    async function fn() {
      const response = await fetch('/api/profile', {
        headers: getHeaders()
      });
      const body = await response.json();
      setProfile(body.profile);
    }

    fn();
  }, []);

  const handleOpenNavMenu = (event) => {
    setAnchorElNav(event.currentTarget);
  };

  const onNavigateTo = (pageName) => {
    setAnchorElNav(null);
    navigate(`/main/${pageName}`);
  };

  const onClose = () => {
    setAnchorElNav(null);
  };

  return (
    <AppBar position="static">
      <Container maxWidth="xl">
        <Toolbar disableGutters>
          <img className="LogoImage" src={logoPng} alt="Logo air conditioner" width="50" height="auto"></img>
          <Box sx={{ flexGrow: 1, display: { xs: 'flex', md: 'none' } }}>
            <IconButton
              size="large"
              aria-label="account of current user"
              aria-controls="menu-appbar"
              aria-haspopup="true"
              onClick={handleOpenNavMenu}
              color="inherit"
            >
              <MenuIcon />
            </IconButton>
            <Menu
              id="menu-appbar"
              anchorEl={anchorElNav}
              anchorOrigin={{
                vertical: 'bottom',
                horizontal: 'left',
              }}
              keepMounted
              transformOrigin={{
                vertical: 'top',
                horizontal: 'left',
              }}
              open={Boolean(anchorElNav)}
              onClose={onClose}
              sx={{
                display: { xs: 'block', md: 'none' },
              }}
            >
              <MenuItem key="home" onClick={() => onNavigateTo('homes')}>
                <Typography textAlign="center">HOMES</Typography>
              </MenuItem>
              <MenuItem key="devices" onClick={() => onNavigateTo('devices')}>
                <Typography textAlign="center">DEVICES</Typography>
              </MenuItem>
            </Menu>
          </Box>
          <Box className="BoxContainer"
               sx={{ flexGrow: 1, display: { xs: 'none', md: 'flex' } }}>
            <Button key="home"
                    onClick={() => onNavigateTo('homes')}
                    sx={{ my: 2, color: 'white', display: 'block' }}>
              HOMES
            </Button>
            <Button key="devices"
                    onClick={() => onNavigateTo('devices')}
                    sx={{ my: 2, color: 'white', display: 'block' }}>
              DEVICES
            </Button>
          </Box>

          <Box sx={{ flexGrow: 0 }}>
            <IconButton onClick={showProfile} sx={{ p: 0 }}>
              <Avatar alt="profile icon" src={profile.github?.avatarURL}/>
            </IconButton>
          </Box>
        </Toolbar>
      </Container>
    </AppBar>
  );
}
