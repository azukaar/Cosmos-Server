import React, { useState } from 'react';
import { Button, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, Stack } from '@mui/material';
import { Alert } from '@mui/material';
import RouteManagement from '../config/routes/routeman';
import { ValidateRoute, getFaviconURL, sanitizeRoute, getContainersRoutes, getHostnameFromName } from '../../utils/routes';
import * as API from '../../api';
import { useTranslation } from 'react-i18next';

const ExposeModal = ({ openModal, setOpenModal, config, updateRoutes, container }) => {
  const { t } = useTranslation();
  const [submitErrors, setSubmitErrors] = useState([]);
  const [newRoute, setNewRoute] = useState(null);

  let containerName = openModal && (openModal.Names[0]);

  const hasCosmosNetwork = () => {
    return container && container.Labels["cosmos-force-network-secured"] === "true";
  }
  
  return <Dialog open={openModal} onClose={() => setOpenModal(false)}>
    <DialogTitle>{t('mgmt.servApp.container.urls.exposeTitle')}</DialogTitle>
        {openModal && <>
        <DialogContent>
            <DialogContentText>
              <Stack spacing={2}>
                <div>
                  {t('mgmt.servApp.container.urls.exposeText')}
                </div>
                <div>
                    <RouteManagement TargetContainer={openModal} 
                      routeConfig={{
                        Target: "http://"+containerName.replace('/', '') + ":",
                        Mode: "SERVAPP",
                        Name: containerName.replace('/', ''),
                        Description: "Expose " + containerName.replace('/', '') + " to the internet",
                        UseHost: true,
                        Host: getHostnameFromName(containerName, null, config),
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
                      config={config}
                      routeNames={config.HTTPConfig.ProxyConfig.Routes.map((r) => r.Name)}
                      setRouteConfig={(_newRoute) => {
                        setNewRoute(sanitizeRoute(_newRoute));
                      }}
                      up={() => {}}
                      down={() => {}}
                      deleteRoute={() => {}}
                      noControls
                      lockTarget
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
            <Button onClick={() => setOpenModal(false)}>{t('global.cancelAction')}</Button>
            <Button onClick={() => {
              let errors = ValidateRoute(newRoute, config);
              if (errors && errors.length > 0) {
                errors = errors.map((err) => {
                  return `${err}`;
                });
                setSubmitErrors(errors);
                return true;
              } else {
                setSubmitErrors([]);
                updateRoutes(newRoute);
              }
              
            }}>{t('global.confirmAction')}</Button>
        </DialogActions>
    </>}
  </Dialog>
};

export default ExposeModal;