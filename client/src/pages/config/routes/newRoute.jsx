// material-ui
import { AppstoreAddOutlined, PlusCircleOutlined, ReloadOutlined, SearchOutlined, SettingOutlined } from '@ant-design/icons';
import { Alert, Badge, Button, Card, Checkbox, Chip, CircularProgress, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, Divider, Input, InputAdornment, TextField, Tooltip, Typography } from '@mui/material';
import Grid2 from '@mui/material/Unstable_Grid2/Grid2';
import { Stack } from '@mui/system';
import { useEffect, useState } from 'react';
import Paper from '@mui/material/Paper';
import { styled } from '@mui/material/styles';

import * as API from '../../../api';
import RestartModal from '../../config/users/restart';
import RouteManagement from '../../config/routes/routeman';
import { ValidateRoute, getFaviconURL, sanitizeRoute } from '../../../utils/routes';
import HostChip from '../../../components/hostChip';
import { useTranslation } from 'react-i18next';


const NewRouteCreate = ({ openNewModal, setOpenNewModal, config }) => {
  const { t } = useTranslation();
  const [openRestartModal, setOpenRestartModal] = useState(false);
  const [submitErrors, setSubmitErrors] = useState([]);
  const [newRoute, setNewRoute] = useState(null);

  function addRoute() {
    return API.config.addRoute(newRoute).then((res) => {
      setOpenNewModal(false);
      setOpenRestartModal(true);
    });
  }

  const routes = config.HTTPConfig.ProxyConfig.Routes || [];

  return <>
  <RestartModal openModal={openRestartModal} setOpenModal={setOpenRestartModal} config={config} newRoute />
  <Dialog open={openNewModal} onClose={() => setOpenNewModal(false)}>
    <DialogTitle>{t('mgmt.urls.edit.newUrlTitle')}</DialogTitle>
        {openNewModal && <>
        <DialogContent>
            <DialogContentText>
              <Stack spacing={2}>
                <div>
                    <RouteManagement
                      routeConfig={{
                        Target: "",
                        Mode: "SERVAPP",
                        Name: "New Route",
                        Description: "New Route",
                        UseHost: false,
                        Host: "",
                        UsePathPrefix: false,
                        PathPrefix: '',
                        CORSOrigin: '',
                        StripPathPrefix: false,
                        AuthEnabled: false,
                        Timeout: 14400000,
                        ThrottlePerMinute: 10000,
                        BlockCommonBots: true,
                        SmartShield: {
                          Enabled: true,
                        }
                      }} 
                      routeNames={routes.map((r) => r.Name)}
                      setRouteConfig={(_newRoute) => {
                        setNewRoute(sanitizeRoute(_newRoute));
                      }}
                      up={() => {}}
                      down={() => {}}
                      deleteRoute={() => {}}
                      noControls
                      config={config}
                    />
                </div>
              </Stack>
            </DialogContentText>
        </DialogContent>
        <DialogActions>
            {submitErrors && submitErrors.length > 0 && <Stack spacing={2} direction={"column"}>
              <Alert severity="error">{submitErrors.map((err) => {
                  return <div>{err}</div>
                })}</Alert>
            </Stack>}
            <Button onClick={() => setOpenNewModal(false)}>{t('global.cancelAction')}</Button>
            <Button onClick={() => {
              let errors = ValidateRoute(newRoute, config);
              if (errors && errors.length > 0) {
                setSubmitErrors(errors);
              } else {
                setSubmitErrors([]);
                addRoute();
              }
              
            }}>{t('global.confirmAction')}</Button>
        </DialogActions>
    </>}
  </Dialog>
  </>;
}

export default NewRouteCreate;