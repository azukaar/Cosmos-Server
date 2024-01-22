import React from 'react';
import { Alert, Checkbox, Chip, CircularProgress, Stack, Typography, useMediaQuery } from '@mui/material';
import MainCard from '../../../components/MainCard';
import { ContainerOutlined, DashboardOutlined, DesktopOutlined, InfoCircleOutlined, NodeExpandOutlined, PlayCircleOutlined, PlusCircleOutlined, SafetyCertificateOutlined, SettingOutlined } from '@ant-design/icons';
import { getFaviconURL, getContainersRoutes } from '../../../utils/routes';
import HostChip from '../../../components/hostChip';
import ExposeModal from '../exposeModal';
import * as API from '../../../api';
import RestartModal from '../../config/users/restart';
import GetActions from '../actionBar';
import { ServAppIcon } from '../../../utils/servapp-icon';
import MiniPlotComponent from '../../dashboard/components/mini-plot';
import UploadButtons from '../../../components/fileUpload';
import NewDockerService from './newService';

const info = {
  backgroundColor: 'rgba(0, 0, 0, 0.1)',
  padding: '10px',
  borderRadius: '5px',
}

const ContainerComposeEdit = ({ containerInfo, config, refresh, updatesAvailable, selfName }) => {
  const isMobile = useMediaQuery((theme) => theme.breakpoints.down('sm'));
  const [openModal, setOpenModal] = React.useState(false);
  const [openRestartModal, setOpenRestartModal] = React.useState(false);
  const [isUpdating, setIsUpdating] = React.useState(false);
  
  const { Name, Config, NetworkSettings, State } = containerInfo;
  const Image = Config.Image;
  const IPAddress = NetworkSettings.Networks?.[Object.keys(NetworkSettings.Networks)[0]]?.IPAddress;
  const Health = State.Health;
  const healthStatus = Health ? Health.Status : 'Healthy';
  const healthIconColor = healthStatus === 'Healthy' ? 'green' : 'red';
  const routes = getContainersRoutes(config, Name.replace('/', ''));

  const [exportedCompose, setExportedCompose] = React.useState(null);

  React.useEffect(() => {
    API.docker.exportContainer(Name.replace('/', '')).then((res) => {
      setExportedCompose({
        services: {
          [Name.replace('/', '')]: res.data
        }
      });
    });
  }, [Name]);

  let refreshAll = refresh ? (() => refresh().then(() => {
    setIsUpdating(false);
  })) : (() => {setIsUpdating(false);});

  return (
    <div style={{ maxWidth: '1000px', width: '100%' }}>
      {exportedCompose && <NewDockerService edit service={exportedCompose} refresh={refreshAll} />}  
    </div>
  );
};

export default ContainerComposeEdit;