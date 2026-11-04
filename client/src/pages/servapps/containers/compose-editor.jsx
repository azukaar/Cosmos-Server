import React from 'react';
import { Alert, FormControlLabel, Stack, Switch, Typography, useMediaQuery } from '@mui/material';
import { useTheme } from '@mui/material/styles';
import MainCard from '../../../components/MainCard';
import * as API from '../../../api';
import NewDockerService from './newService';
import { PERM_CREDENTIALS_READ } from '../../../utils/permissions';
import { useClientInfos } from '../../../utils/hooks';
import { useTranslation } from 'react-i18next';

const ContainerComposeEdit = ({ containerInfo, config, refresh, updatesAvailable, selfName }) => {
  const theme = useTheme();
  const isMobile = useMediaQuery((theme) => theme.breakpoints.down('sm'));
  const [isUpdating, setIsUpdating] = React.useState(false);
  const { hasPermission, hasRolePermission } = useClientInfos();
  const { t } = useTranslation();

  const { Name } = containerInfo;

  const [exportedCompose, setExportedCompose] = React.useState(null);
  const [savedComments, setSavedComments] = React.useState(null);
  // false = show the config the user actually set (container vs image-config
  // diff: only explicitly-set values). true = show the live Docker runtime
  // state (everything Docker resolved, including image defaults).
  const [showRuntime, setShowRuntime] = React.useState(false);

  React.useEffect(() => {
    if (!hasPermission(PERM_CREDENTIALS_READ)) return;
    API.docker.exportContainer(Name.replace('/', ''), showRuntime ? 'runtime' : 'initial').then((res) => {
      setExportedCompose({
        services: {
          [Name.replace('/', '')]: res.data
        }
      });
      setSavedComments(res.comments || null);
    });
  }, [Name, showRuntime]);

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
      <Stack direction="row" spacing={2} alignItems="center" justifyContent="space-between" sx={{ mb: 1 }}>
        <Typography variant="caption" color="text.secondary">
          {showRuntime
            ? 'Showing live runtime values (including image defaults)'
            : 'Showing values that were explicitly set (image defaults excluded)'}
        </Typography>
        <FormControlLabel
          control={
            <Switch
              checked={showRuntime}
              onChange={(e) => setShowRuntime(e.target.checked)}
              size="small"
              sx={{
                '& .MuiSwitch-switchBase.Mui-checked': {
                  color: theme.palette.primary.main,
                  '&:hover': { backgroundColor: theme.palette.primary.main + '22' },
                },
                '& .MuiSwitch-switchBase.Mui-checked + .MuiSwitch-track': {
                  backgroundColor: theme.palette.primary.main,
                  opacity: 1,
                },
                '& .MuiSwitch-track': {
                  backgroundColor: theme.palette.mode === 'dark'
                    ? theme.palette.grey[600]
                    : theme.palette.grey[400],
                  opacity: 1,
                },
              }}
            />
          }
          label={
            <Typography variant="body2" sx={{ color: theme.palette.text.primary }}>
              Show runtime values
            </Typography>
          }
          sx={{ '& .MuiFormControlLabel-label': { color: theme.palette.text.primary } }}
        />
      </Stack>
      {exportedCompose && <NewDockerService edit service={exportedCompose} refresh={refreshAll} comments={savedComments} />}
    </div>
  );
};

export default ContainerComposeEdit;
