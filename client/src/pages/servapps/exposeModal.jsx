import React, { useState } from 'react';
import { Button, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, Stack } from '@mui/material';
import { Alert } from '@mui/material';
import RouteManagement from '../config/routes/routeman';
import { ValidateRoute, getFaviconURL, sanitizeRoute, getContainersRoutes } from '../../utils/routes';
import * as API from '../../api';

const getHostnameFromName = (name) => {
  return name.replace('/', '').replace(/_/g, '-').replace(/[^a-zA-Z0-9-]/g, '').toLowerCase().replace(/\s/g, '-') + '.' + window.location.origin.split('://')[1]
}

const ExposeModal = ({ openModal, setOpenModal, config, updateRoutes, container }) => {
  const [submitErrors, setSubmitErrors] = useState([]);
  const [newRoute, setNewRoute] = useState(null);

  let containerName = openModal && (openModal.Names[0]);

  const hasCosmosNetwork = () => {
    return container && container.Labels["cosmos-force-network-secured"] === "true";
  }
  
  return <Dialog open={openModal} onClose={() => setOpenModal(false)}>
    <DialogTitle>Expose ServApp</DialogTitle>
        {openModal && <>
        <DialogContent>
            <DialogContentText>
              <Stack spacing={2}>
                <div>
                  Welcome to the URL Wizard. This interface will help you expose your ServApp securely to the internet by creating a new URL.
                </div>
                <div>
                    {openModal && !hasCosmosNetwork(containerName) && 
                      <Alert severity="warning">
                        This ServApp does not use the "Force Secure" option, 
                        so the hostname might not be accessible.
                        The easiest way to fix this is to check the box "Force Secure Network" or manually create a hostname and sub-network in Docker.
                      </Alert>}
                </div>
                <div>
                    <RouteManagement TargetContainer={openModal} 
                      routeConfig={{
                        Target: "http://"+containerName.replace('/', '') + ":",
                        Mode: "SERVAPP",
                        Name: containerName.replace('/', ''),
                        Description: "Expose " + containerName.replace('/', '') + " to the internet",
                        UseHost: true,
                        Host: getHostnameFromName(containerName),
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
            <Button onClick={() => setOpenModal(false)}>Cancel</Button>
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
              
            }}>Confirm</Button>
        </DialogActions>
    </>}
  </Dialog>
};

export default ExposeModal;