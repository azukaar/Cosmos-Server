import React from 'react';
import { Box, IconButton, LinearProgress, Stack, Tooltip, useMediaQuery } from '@mui/material';
import { CheckCircleOutlined, CloseSquareOutlined, DeleteOutlined, PauseCircleOutlined, PlaySquareOutlined, ReloadOutlined, RollbackOutlined, StopOutlined, UpCircleOutlined } from '@ant-design/icons';
import * as API from '../../api';
import LogsInModal from '../../components/logsInModal';

const GetActions = ({
  Id,
  state,
  image,
  refreshServApps,
  setIsUpdatingId,
  updateAvailable
}) => {
  const [confirmDelete, setConfirmDelete] = React.useState(false);
  const isMiniMobile = useMediaQuery((theme) => theme.breakpoints.down('xsm'));
  const [pullRequest, setPullRequest] = React.useState(null);
  const [isUpdating, setIsUpdating] = React.useState(false);
  
  const doTo = (action) => {
    setIsUpdating(true);

    if(action === 'update') {
      setPullRequest(() => ((cb) => {
        API.docker.updateContainerImage(Id, cb, true)
      }));
      return;
    }

    setIsUpdatingId(Id, true);
    return API.docker.manageContainer(Id, action).then((res) => {
      setIsUpdating(false);
      refreshServApps();
    }).catch((err) => {
      setIsUpdating(false);
      refreshServApps();
    });
  };

  let actions = [
    {
      t: 'Update Available, Click to Update',
      if: ['update_available'],
      e: <IconButton className="shinyButton" color='primary' onClick={() => {doTo('update')}} size={isMiniMobile ? 'medium' : 'large'}>
        <UpCircleOutlined />
      </IconButton>
    },
    {
      t: 'No Update Available. Click to Force Pull',
      if: ['update_not_available'],
      e: <IconButton onClick={() => {doTo('update')}} size={isMiniMobile ? 'medium' : 'large'}>
        <UpCircleOutlined />
      </IconButton>
    },
    {
      t: 'Start',
      if: ['exited', 'created'],
      e: <IconButton onClick={() => {doTo('start')}} size={isMiniMobile ? 'medium' : 'large'}>
        <PlaySquareOutlined />
      </IconButton>
    },
    {
      t: 'Unpause',
      if: ['paused'],
      e: <IconButton onClick={() => {doTo('unpause')}} size={isMiniMobile ? 'medium' : 'large'}>
        <PlaySquareOutlined />
      </IconButton>
    },
    {
      t: 'Pause',
      if: ['running'],
      e: <IconButton onClick={() => {doTo('pause')}} size={isMiniMobile ? 'medium' : 'large'}>
        <PauseCircleOutlined />
      </IconButton>
    },
    {
      t: 'Stop',
      if: ['paused', 'restarting', 'running'],
      e: <IconButton onClick={() => {doTo('stop')}} size={isMiniMobile ? 'medium' : 'large'} variant="outlined">
        <StopOutlined />
      </IconButton>
    },
    {
      t: 'Restart',
      if: ['exited', 'running', 'paused', 'created', 'restarting'],
      e: <IconButton onClick={() => doTo('restart')} size={isMiniMobile ? 'medium' : 'large'}>
        <ReloadOutlined />
      </IconButton>
    },
    {
      t: 'Re-create',
      if: ['exited', 'running', 'paused', 'created', 'restarting'],
      e: <IconButton onClick={() => doTo('recreate')} color="error" size={isMiniMobile ? 'medium' : 'large'}>
        <RollbackOutlined />
      </IconButton>
    },
    {
      t: 'Kill',
      if: ['running', 'paused', 'created', 'restarting'],
      e: <IconButton onClick={() => doTo('kill')} color="error" size={isMiniMobile ? 'medium' : 'large'}>
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

  return <>
    <LogsInModal
      request={pullRequest}
      title="Updating ServeApp..."
      OnSuccess={() => {
        refreshServApps();
      }}
    />
    
    {!isUpdating && actions.filter((action) => {
      return action.if.includes(state) || (updateAvailable && action.if.includes('update_available')) || (!updateAvailable && action.if.includes('update_not_available'));
    }).map((action) => {
      return <Tooltip title={action.t}>{action.e}</Tooltip>
    })}

    {isUpdating && <Stack sx={{
      width: '100%', height: '44px',
    }}
    justifyContent={'center'} 
    ><LinearProgress /></Stack>}
  </>
}

export default GetActions;