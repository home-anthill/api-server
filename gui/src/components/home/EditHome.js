import React from 'react';
import { useLocation, useNavigate } from 'react-router-dom';

import { useForm, Controller, useFieldArray } from 'react-hook-form';

import { Typography, Button, TextField, FormControl, Stack, IconButton } from '@mui/material';
import CheckIcon from '@mui/icons-material/Check';
import DeleteIcon from '@mui/icons-material/Delete';

import './EditHome.css';

import { deleteApi, postApi, putApi } from '../../apis/api';

export default function EditHome() {
  const { state } = useLocation();
  const homeInput = state.home;

  const {handleSubmit, control, getValues} = useForm({
    defaultValues: {
      nameInput: homeInput.name,
      locationInput: homeInput.location
    }
  });

  const roomsForm = useForm({
    defaultValues: {
      rooms: homeInput.rooms
    }
  });
  const { fields, append, remove } = useFieldArray({
    control: roomsForm.control,
    name: "rooms"
  });

  const navigate = useNavigate();

  const onAddHome = async () => {
    const values = getValues();
    try {
      await putApi(`/api/homes/${homeInput.id}`, {
        name: values.nameInput,
        location: values.locationInput,
        // cannot change room with this api
      });
      // navigate back
      navigate(-1);
    } catch (err) {
      console.error('Cannot add a new home');
    }
  }

  async function onRemoveRoom(index) {
    const room = roomsForm.getValues().rooms[index];
    // remove room from array
    remove(index);
    // in case you are creating a room, and you decide to remove it before adding it to the server
    if (!room.id) {
      return;
    }
    try {
      // remove remove from server
      await deleteApi(`/api/homes/${homeInput.id}/rooms/${room.id}`);
      // navigate back
      navigate(-1);
    } catch (err) {
      console.error('Cannot add a new home');
    }
  }

  async function onSaveRoom(e) {
    e.preventDefault();
    const rooms = roomsForm.getValues().rooms;
    try {
      for (let room of rooms) {
        if (room.id) {
          console.log('ADDING NEW ROOM', room);
          await putApi(`/api/homes/${homeInput.id}/rooms/${room.id}`, {
            name: room.name,
            floor: +room.floor
            // cannot change room with this api
          });
        } else {
          console.log('UPDATING ROOM', room);
          await postApi(`/api/homes/${homeInput.id}/rooms`, {
            name: room.name,
            floor: +room.floor
            // cannot change room with this api
          });
        }
      }
      // navigate back
      navigate(-1);
    } catch (err) {
      console.error('Cannot add a new home');
    }
  }

  return (
    <div className="EditHome">
      <Typography variant="h2" component="h1">
        Edit Home
      </Typography>
      <div className="EditHomeContainer">
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
                    left: 15
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
      </div>
      <Button variant="outlined" onClick={() => onAddHome()}>Save Home</Button>

      <div className="EditHomeDivider"></div>

      <Typography variant="h2" component="h1">
        Rooms
      </Typography>
      <Stack
        direction="column"
        justifyContent="center"
        alignItems="center"
        spacing={1}
        sx={{
          marginTop: '30px'
        }}
      >
        {fields.map((item, index) => (
          <form className="Room" key={`room-${index}`}>
            <FormControl>
              <Controller
                render={({field}) =>
                  <TextField
                    variant="standard"
                    label="Name"
                    {...field} />
                }
                name={`rooms.${index}.name`}
                rules={{required: true, maxLength: 15}}
                control={roomsForm.control}
              />
            </FormControl>
            <FormControl>
              <Controller
                render={({field}) =>
                  <TextField
                    sx={{
                      left: 15
                    }}
                    variant="standard"
                    label="Floor"
                    inputProps={{ inputMode: 'numeric', pattern: '[0-9]*' }}
                    {...field} />
                }
                name={`rooms.${index}.floor`}
                rules={{required: true}}
                control={roomsForm.control}
              />
            </FormControl>
            <IconButton
              aria-label="save"
              onClick={onSaveRoom}>
              <CheckIcon />
            </IconButton>
            <IconButton
              aria-label="delete"
              onClick={() => onRemoveRoom(index)}>
              <DeleteIcon />
            </IconButton>
          </form>
      ))}
      </Stack>
      <br/>
      <br/>
      <Button variant="standard" onClick={() => {
        append({ name: '', floor: 0 });
      }}>(+ add room)</Button>

      <br/><br/>
    </div>
  )
}

