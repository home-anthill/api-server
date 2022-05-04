import React, { useState } from 'react';

import { Card, CardActions, CardContent, Typography, IconButton, Collapse, styled } from '@mui/material';
import EditIcon from '@mui/icons-material/Edit';
import DeleteIcon from '@mui/icons-material/Delete';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';

export default function Home({home, onEdit, onDelete}) {
  const [expanded, setExpanded] = useState(false);

  const ExpandMore = styled((props) => {
    const { expand, ...other } = props;
    return <IconButton {...other} />;
  })(({ theme, expand }) => ({
    transform: !expand ? 'rotate(0deg)' : 'rotate(180deg)',
    marginLeft: 'auto',
    transition: theme.transitions.create('transform', {
      duration: theme.transitions.duration.shortest,
    }),
  }));

  const handleExpandClick = () => {
    setExpanded(!expanded);
  };

  return (
    <Card variant="outlined"
          sx={{
            margin: "12px",
            minWidth: "250px"
          }}>
      <CardContent>
        <Typography variant="h5" component="div">
          {home.name}
        </Typography>
        <Typography sx={{ fontSize: 14 }} color="text.secondary" gutterBottom>
          {home.location}
        </Typography>
      </CardContent>
      <CardActions>
        <IconButton aria-label="edit" onClick={() => onEdit(home)}>
          <EditIcon />
        </IconButton>
        <IconButton aria-label="delete" onClick={() => onDelete(home)}>
          <DeleteIcon />
        </IconButton>
        {home.rooms && home.rooms.length > 0 &&
          <ExpandMore
            expand={expanded}
            onClick={handleExpandClick}
            aria-expanded={expanded}
            aria-label="show rooms"><ExpandMoreIcon/>
          </ExpandMore>
        }
      </CardActions>
      <Collapse in={expanded} timeout="auto" unmountOnExit>
        <CardContent>
          {home.rooms && home.rooms.map(room => (
            <Typography paragraph key={room.id}>{room.name} @ {room.floor}</Typography>
          ))}
        </CardContent>
      </Collapse>
    </Card>
  )
}

