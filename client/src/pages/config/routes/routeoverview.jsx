import * as React from 'react';
import MainCard from '../../../components/MainCard';
import RestartModal from '../users/restart';
import { Chip, Divider, Stack, useMediaQuery } from '@mui/material';
import HostChip from '../../../components/hostChip';
import { RouteMode, RouteSecurity } from '../../../components/routeComponents';
import { getFaviconURL } from '../../../utils/routes';
import * as API from '../../../api';
import { CheckOutlined, ClockCircleOutlined, DashboardOutlined, DeleteOutlined, DownOutlined, LockOutlined, UpOutlined } from "@ant-design/icons";
import IsLoggedIn from '../../../isLoggedIn';
import { redirectToLocal } from '../../../utils/indexs';

const info = {
  backgroundColor: 'rgba(0, 0, 0, 0.1)',
  padding: '10px',
  borderRadius: '5px',
}

const RouteOverview = ({ routeConfig }) => {
  const isMobile = useMediaQuery((theme) => theme.breakpoints.down('sm'));
  const [confirmDelete, setConfirmDelete] = React.useState(false);

  function deleteRoute(event) {
    event.stopPropagation();
    API.config.deleteRoute(routeConfig.Name).then(() => {
      redirectToLocal('/cosmos-ui/config-url');
    });
  }

  return <div style={{ maxWidth: '1000px', width: '100%'}}>
    {routeConfig && <>
      <MainCard name={routeConfig.Name} title={<div>
        {routeConfig.Name} &nbsp;&nbsp;
        {!confirmDelete && (<Chip label={<DeleteOutlined />} onClick={() => setConfirmDelete(true)}/>)}
        {confirmDelete && (<Chip label={<CheckOutlined />} color="error" onClick={(event) => deleteRoute(event)}/>)}
      </div>}>
        <Stack spacing={2} direction={isMobile ? 'column' : 'row'} alignItems={isMobile ? 'center' : 'flex-start'}>
          <div>
            <img className="loading-image" alt="" src={getFaviconURL(routeConfig)} width="128px" />
          </div>
          <Stack spacing={2} >
            <strong>Description</strong>
            <div style={info}>{routeConfig.Description}</div>
            <strong>URL</strong>
            <div><HostChip route={routeConfig} /></div>
            <strong>Target</strong>
            <div><RouteMode route={routeConfig} /> <Chip label={routeConfig.Target} /></div>
            <strong>Security</strong>
            <div><RouteSecurity route={routeConfig} /></div>
          </Stack>
        </Stack>
      </MainCard>
    </>}
  </div>;
}

export default RouteOverview;