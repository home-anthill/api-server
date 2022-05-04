import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';

import { Typography, Fab, Dialog, DialogTitle, Button, DialogContent, DialogContentText, TextField, DialogActions, FormControl } from '@mui/material';
import AddIcon from '@mui/icons-material/Add';
import PropTypes from 'prop-types';

import axios from 'axios';

import './Homes.css';

import Home from '../../shared/Home';
import useHomes from '../../apis/useHomes';

export default function Homes() {
  const [open, setOpen] = useState(false);
  const [homes, setHomes] = useState([]);
  const {
    data: homesData,
    loading: homesLoading,
    error: homesError
  } = useHomes(homes);

  const navigate = useNavigate();

  function editHome(home) {
    navigate(`/main/homes/${home.id}/edit`, {state: {home}});
  }

  async function deleteHome(home) {
    let token = localStorage.getItem('token');
    let headers = {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer ' + token
    };
    try {
      await axios.delete(`/api/homes/${home.id}`, {
        headers
      })
      setHomes([]);
    } catch (err) {
      console.error(`Cannot delete home with id = ${home.id}`);
    }
  }

  const handleOpen = () => {
    setOpen(true);
  };

  const handleClose = (value) => {
    if (value) {
      setHomes([]);
    }
    setOpen(false);
  };

  return (
    <div className="Homes">
      <NewHomeDialog
        open={open}
        onClose={handleClose}/>
      <Typography variant="h2" component="h1">
        Homes
      </Typography>
      <div className="HomesContainer">
        {homesError ? (
          <div className="error">
            Something went wrong:
            <div className="error-contents">
              {homesError.message}
            </div>
          </div>
        ) : homesLoading ? (
          <div className="loading">Loading...</div>
        ) : homesData && homesData.length ? (
          <>
            {homesData.map((home) => (
              <Home key={home.id}
                    home={home}
                    onEdit={editHome}
                    onDelete={() => deleteHome(home)}>
              </Home>
            ))}
          </>
        ) : (
          'No homes'
        )}
      </div>
      <Fab color="primary"
           sx={{
             position: 'absolute',
             bottom: 32,
             right: 32
           }}
           aria-label="add"
           onClick={handleOpen}>
        <AddIcon/>
      </Fab>
    </div>
  )
}

function NewHomeDialog({onClose, open}) {
  const [home, setHome] = useState({name: '', location: ''});

  const handleClose = () => {
    onClose(false);
  };

  const handleAdd = (value) => {
    onClose(value);
  };

  const submit = async e => {
    e.preventDefault(); // prevent default submit

    let token = localStorage.getItem('token');
    let headers = {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer ' + token
    };
    try {
      await
        axios.post(`/api/homes`, {
          name: home.name,
          location: home.location,
          rooms: []
        }, {
          headers
        });
      handleAdd(true);
    } catch (err) {
      console.error('Cannot add a new home');
      handleAdd(false);
    }
  }

  const onChangeHomeName = value => {
    setHome(prevState => ({
      ...prevState,
      name: value
    }));
  }
  const onChangeHomeLocation = value => {
    setHome(prevState => ({
      ...prevState,
      location: value
    }));
  }

  return (
    <Dialog open={open} onClose={handleClose}>
      <DialogTitle>Create a new home</DialogTitle>
      <DialogContent>
        <DialogContentText>
          To subscribe to this website, please enter your email address here. We
          will send updates occasionally.
        </DialogContentText>
        <form onSubmit={submit}>
          <FormControl>
            <TextField
              id="name-input"
              variant="outlined"
              required
              value={home.name}
              onChange={e => onChangeHomeName(e.target.value)}
              label="Name"/>
          </FormControl>
          <FormControl>
            <TextField
              id="location-input"
              variant="outlined"
              required
              value={home.location}
              onChange={e => onChangeHomeLocation(e.target.value)}
              label="Location"/>
          </FormControl>
        </form>
      </DialogContent>
      <DialogActions>
        <Button onClick={() => handleClose()}>Cancel</Button>
        <Button onClick={submit}>Add</Button>
      </DialogActions>
    </Dialog>
  );
}

NewHomeDialog.propTypes = {
  onClose: PropTypes.func.isRequired,
  open: PropTypes.bool.isRequired
};
