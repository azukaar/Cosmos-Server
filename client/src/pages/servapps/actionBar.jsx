import React from 'react';
import { IconButton, Tooltip } from '@mui/material';
import { CheckCircleOutlined, CloseSquareOutlined, DeleteOutlined, PauseCircleOutlined, PlaySquareOutlined, ReloadOutlined, RollbackOutlined, StopOutlined, UpCircleOutlined } from '@ant-design/icons';
import * as API from '../../api';

const GetActions = ({
  Id,
  state,
  refreshServeApps,
  setIsUpdatingId
}) => {
  const [confirmDelete, setConfirmDelete] = React.useState(false);

  const doTo = (action) => {
    setIsUpdatingId(Id, true);
    API.docker.manageContainer(Id, action).then((res) => {
      refreshServeApps();
    });
  };

  let actions = [
    {
      t: 'Update Available',
      if: ['update_available'],
      e: <IconButton className="shinyButton" color='primary' onClick={() => {doTo('update')}} size='large'>
        <UpCircleOutlined />
      </IconButton>
    },
    {
      t: 'Start',
      if: ['exited', 'created'],
      e: <IconButton onClick={() => {doTo('start')}} size='large'>
        <PlaySquareOutlined />
      </IconButton>
    },
    {
      t: 'Unpause',
      if: ['paused'],
      e: <IconButton onClick={() => {doTo('unpause')}} size='large'>
        <PlaySquareOutlined />
      </IconButton>
    },
    {
      t: 'Pause',
      if: ['running'],
      e: <IconButton onClick={() => {doTo('pause')}} size='large'>
        <PauseCircleOutlined />
      </IconButton>
    },
    {
      t: 'Stop',
      if: ['paused', 'restarting', 'running'],
      e: <IconButton onClick={() => {doTo('stop')}} size='large' variant="outlined">
        <StopOutlined />
      </IconButton>
    },
    {
      t: 'Restart',
      if: ['exited', 'running', 'paused', 'created', 'restarting'],
      e: <IconButton onClick={() => doTo('restart')} size='large'>
        <ReloadOutlined />
      </IconButton>
    },
    {
      t: 'Re-create',
      if: ['exited', 'running', 'paused', 'created', 'restarting'],
      e: <IconButton onClick={() => doTo('recreate')} color="error" size='large'>
        <RollbackOutlined />
      </IconButton>
    },
    {
      t: 'Kill',
      if: ['running', 'paused', 'created', 'restarting'],
      e: <IconButton onClick={() => doTo('kill')} color="error" size='large'>
        <CloseSquareOutlined />
      </IconButton>
    },
    {
      t: 'Delete',
      if: ['exited', 'created'],
      e: <IconButton onClick={() => {
        if(confirmDelete) doTo('remove')
        else setConfirmDelete(true);
      }} color="error" size='large'>
        {confirmDelete ?  <CheckCircleOutlined />
        : <DeleteOutlined />}
      </IconButton>
    }
  ];

  return actions.filter((action) => {
    let updateAvailable = false;
    return action.if.includes(state) ?? (updateAvailable && action.if.includes('update_available'));
  }).map((action) => {
    return <Tooltip title={action.t}>{action.e}</Tooltip>
  });
}

export default GetActions;