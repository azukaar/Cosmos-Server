import React from 'react';
import { Alert, Checkbox, Chip, CircularProgress, Stack, Typography, useMediaQuery } from '@mui/material';
import MainCard from '../../../components/MainCard';
import * as API from '../../../api';
import NewDockerService from './newService';
import { PERM_CREDENTIALS_READ } from '../../../utils/permissions';
import { useClientInfos } from '../../../utils/hooks';
import { useTranslation } from 'react-i18next';

const ContainerComposeEdit = ({ containerInfo, config, refresh, updatesAvailable, selfName }) => {
  const isMobile = useMediaQuery((theme) => theme.breakpoints.down('sm'));
  const [isUpdating, setIsUpdating] = React.useState(false);
  const { hasPermission, hasRolePermission } = useClientInfos();
  const { t } = useTranslation();

  const { Name } = containerInfo;

  const [exportedCompose, setExportedCompose] = React.useState(null);

  React.useEffect(() => {
    if (!hasPermission(PERM_CREDENTIALS_READ)) return;
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

  if (!hasPermission(PERM_CREDENTIALS_READ)) {
    return (
      <div style={{ maxWidth: '1000px', width: '100%', margin: 'auto', padding: '20px 0' }}>
        <Alert severity="warning">
          {hasRolePermission(PERM_CREDENTIALS_READ)
            ? t('mgmt.servapps.compose.credentialsRequired')
            : t('mgmt.servapps.compose.credentialsDenied')}
        </Alert>
      </div>
    );
  }

  return (
    <div style={{ maxWidth: '1000px', width: '100%', margin: 'auto' }}>
      {exportedCompose && <NewDockerService edit service={exportedCompose} refresh={refreshAll} />}
    </div>
  );
};

export default ContainerComposeEdit;
