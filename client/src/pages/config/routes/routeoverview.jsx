import * as React from 'react';
import MainCard from '../../../components/MainCard';
import RestartModal from '../users/restart';
import { Checkbox, Chip, Divider, FormControlLabel, Stack, useMediaQuery } from '@mui/material';
import HostChip from '../../../components/hostChip';
import { RouteMode, RouteSecurity } from '../../../components/routeComponents';
import { getFaviconURL } from '../../../utils/routes';
import * as API from '../../../api';
import { CheckOutlined, ClockCircleOutlined, ContainerOutlined, DashboardOutlined, DeleteOutlined, DownOutlined, InfoCircleFilled, InfoCircleOutlined, LockOutlined, NodeExpandOutlined, SafetyCertificateOutlined, UpOutlined } from "@ant-design/icons";
import { redirectToLocal } from '../../../utils/indexs';
import { CosmosCheckbox } from '../users/formShortcuts';
import { Field } from 'formik';
import MiniPlotComponent from '../../dashboard/components/mini-plot';
import ImageWithPlaceholder from '../../../components/imageWithPlaceholder';
import UploadButtons from '../../../components/fileUpload';
import { useTranslation } from 'react-i18next';

const info = {
  backgroundColor: 'rgba(0, 0, 0, 0.1)',
  padding: '10px',
  borderRadius: '5px',
}

const RouteOverview = ({ routeConfig }) => {
  const { t } = useTranslation();
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
            <ImageWithPlaceholder className="loading-image" alt="" src={getFaviconURL(routeConfig)} width="128px" />
          </div>
          <Stack spacing={2} style={{ width: '100%' }}>
            <strong><ContainerOutlined /> {t('global.description')}</strong>
            <div style={info}>{routeConfig.Description}</div>
            <strong><NodeExpandOutlined /> URL</strong>
            <div><HostChip route={routeConfig} /></div>
            <strong><InfoCircleOutlined /> {t('global.target')}</strong>
            <div><RouteMode route={routeConfig} /> <Chip label={routeConfig.Target} /></div>
            <strong><SafetyCertificateOutlined/> {t('global.securityTitle')}</strong>
            <div><RouteSecurity route={routeConfig} /></div>
            <strong><DashboardOutlined/> {t('menu-items.navigation.monitoringTitle')}</strong>
            <div>
              <MiniPlotComponent agglo metrics={[
                "cosmos.proxy.route.success." + routeConfig.Name,
                "cosmos.proxy.route.error." + routeConfig.Name,
              ]} labels={{
                ["cosmos.proxy.route.error." + routeConfig.Name]: t('global.error'), 
                ["cosmos.proxy.route.success." + routeConfig.Name]: t('global.success')
              }}/>
              <MiniPlotComponent agglo metrics={[
                "cosmos.proxy.route.bytes." + routeConfig.Name,
                "cosmos.proxy.route.time." + routeConfig.Name,
              ]} labels={{
                ["cosmos.proxy.route.bytes." + routeConfig.Name]: "Bytes", 
                ["cosmos.proxy.route.time." + routeConfig.Name]: t('global.time')
              }}/>
            </div>
          </Stack>
        </Stack>
      </MainCard>
    </>}
  </div>;
}

export default RouteOverview;