import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';

import { useForm, Controller } from 'react-hook-form';
import PropTypes from 'prop-types';

import { Typography, Fab, Dialog, DialogTitle, Button, DialogContent, TextField, DialogActions, FormControl } from '@mui/material';
import AddIcon from '@mui/icons-material/Add';

import './Homes.css';

import { deleteApi, postApi } from '../../apis/api';
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
    try {
      await deleteApi(`/api/homes/${home.id}`)
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
  const defaultValues = {
    nameInput: '',
    locationInput: ''
  };
  const {handleSubmit, reset, control, getValues} = useForm({defaultValues});

  const handleClose = () => {
    reset();
    onClose(false);
  };

  const handleAdd = (value) => {
    reset();
    onClose(value);
  };

  const onAddHome = async () => {
    const values = getValues();
    try {
      await postApi(`/api/homes`, {
        name: values.nameInput,
        location: values.locationInput,
        rooms: []
      });
      handleAdd(true);
    } catch (err) {
      console.error('Cannot add a new home');
      handleAdd(false);
    }
  }

  return (
    <Dialog open={open} onClose={handleClose}>
      <DialogTitle>Create a new home</DialogTitle>
      <DialogContent>
        <form onSubmit={handleSubmit((data) => onAddHome())} className="form">
          <FormControl>
            <Controller
              render={({field}) =>
                <TextField
                  id="name-input"
                  variant="outlined"
                  label="Name"
                  {...field} />
              }
              name="nameInput"
              rules={{required: true, maxLength: 15}}
              control={control}
            />
          </FormControl>
          <FormControl>
            <Controller
              render={({field}) =>
                <TextField
                  sx={{
                    marginLeft: '15px'
                  }}
                  id="location-input"
                  variant="outlined"
                  label="Location"
                  {...field} />
              }
              name="locationInput"
              rules={{required: true, maxLength: 15}}
              control={control}
            />
          </FormControl>
        </form>
      </DialogContent>
      <DialogActions>
        <Button onClick={() => handleClose()}>Cancel</Button>
        <Button onClick={() => onAddHome()}>Add</Button>
      </DialogActions>
    </Dialog>
  );
}

NewHomeDialog.propTypes = {
  onClose: PropTypes.func.isRequired,
  open: PropTypes.bool.isRequired
};
