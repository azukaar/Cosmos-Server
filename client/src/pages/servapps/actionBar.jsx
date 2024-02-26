import React from 'react';
import { Box, Chip, IconButton, LinearProgress, Stack, Tooltip, useMediaQuery } from '@mui/material';
import { CheckCircleOutlined, CloseSquareOutlined, DeleteOutlined, PauseCircleOutlined, PlaySquareOutlined, ReloadOutlined, RollbackOutlined, StopOutlined, UpCircleOutlined } from '@ant-design/icons';
import * as API from '../../api';
import LogsInModal from '../../components/logsInModal';
import DeleteModal from './deleteModal';

const GetActions = ({
  Id,
  Ids,
  state,
  image,
  refreshServApps,
  setIsUpdatingId,
  updateAvailable,
  isStack,
  containers,
  config,
}) => {
  const [confirmDelete, setConfirmDelete] = React.useState(false);
  const isMiniMobile = useMediaQuery((theme) => theme.breakpoints.down('xsm'));
  const [pullRequest, setPullRequest] = React.useState(null);
  const [isUpdating, setIsUpdating] = React.useState(false);

  
  const doTo = (action) => {
    setIsUpdating(true);

    if(action === 'update') {
      setPullRequest(() => ((cb) => {
        return API.docker.updateContainerImage(Id, cb, true)
      }));
      return;
    }

    return isStack ?
    (() => {
      Promise.all(Ids.map((id) => {
        setIsUpdatingId(id, true);

        return API.docker.manageContainer(id, action)
      })).then((res) => {
        setIsUpdating(false);
        refreshServApps();
      }).catch((err) => {
        setIsUpdating(false);
        refreshServApps();
      })
    })()
    : 
    (() => {
      setIsUpdatingId(Id, true);

      API.docker.manageContainer(Id, action).then((res) => {
        setIsUpdating(false);
        refreshServApps();
      }).catch((err) => {
        setIsUpdating(false);
        refreshServApps();
      });
    })()
  };

  let actions = [
    {
      t: 'Update Available' + (isStack ? ', go the stack details to update' : ', Click to Update'),
      if: ['update_available'],
      es:  <IconButton className="shinyButton" style={{cursor: 'not-allowed'}} color='primary' onClick={()=>{}} size={isMiniMobile ? 'medium' : 'large'}>
      <UpCircleOutlined />
    </IconButton>,
      e: <IconButton className="shinyButton" color='primary' onClick={() => {doTo('update')}} size={isMiniMobile ? 'medium' : 'large'}>
        <UpCircleOutlined />
      </IconButton>
    },
    {
      t: 'No Update Available. Click to Force Pull',
      if: ['update_not_available'],
      hideStack: true,
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
      hideStack: true,
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
      e: <DeleteModal config={config} Ids={Ids} containers={containers} refreshServApps={refreshServApps} setIsUpdatingId={setIsUpdatingId} />
    }
  ];

  return <>
    {pullRequest && <LogsInModal
      request={pullRequest}
      title="Updating ServApp..."
      OnSuccess={() => {
        refreshServApps();
        setPullRequest(null);
      }}
      OnClose={() => {
        setPullRequest(null);
      }}
    />}
    
    {!isUpdating && actions.filter((action) => {
      return action.if.includes(state) || (updateAvailable && action.if.includes('update_available')) || (!updateAvailable && action.if.includes('update_not_available'));
    }).map((action) => {
      return (!isStack || !action.hideStack) && <Tooltip title={action.t}>{isStack ? (action.es ? action.es : action.e) : action.e}</Tooltip>
    })}

    {isUpdating && <Stack sx={{
      width: '100%', height: '44px',
    }}
    justifyContent={'center'} 
    ><LinearProgress /></Stack>}
  </>
}

export default GetActions;