import React from 'react';
import { Box, Stack } from '@mui/material';
import Grid2 from '@mui/material/Unstable_Grid2/Grid2';
import { Link } from 'react-router-dom';
import { getFullOrigin } from '../../utils/routes';
import { ServAppIcon } from '../../utils/servapp-icon';
import { useDragAndDrop } from '@formkit/drag-and-drop/react';

const AppGrid = ({ config, servApps, routes, coStatus, appColor, appBorder, blockStyle }) => {
  const validRoutes = routes ? routes.filter(route => !route.HideFromDashboard) : [];
  const [parentRef, apps] = useDragAndDrop(validRoutes);

  return (
    <Grid2 container spacing={2} ref={parentRef}>
      {config && servApps && apps.map((route) => {
        let skip = route.Mode === 'REDIRECT';
        let containerName;
        let container;
        if (route.Mode === 'SERVAPP') {
          containerName = route.Target.split(':')[1].slice(2);
          container = servApps.find((c) => c.Names.includes('/' + containerName));
          // TOOD: rework, as it prevents users from seeing the apps
          // if (!container || container.State != "running") {
          //     skip = true
          // }
        }

        if (route.HideFromDashboard) skip = true;

        return (
          !skip && coStatus && (coStatus.homepage.Expanded ? (
            <Grid2 xs={12} sm={6} md={4} lg={3} xl={3} xxl={3} key={route.Name}>
              <Box className='app app-hover' style={{ padding: 25, borderRadius: 5, ...appColor, ...appBorder }}>
                <Link to={getFullOrigin(route)} target="_blank" style={{ textDecoration: 'none', ...appColor }}>
                  <Stack direction='row' spacing={2} alignItems='center'>
                    <ServAppIcon container={container} route={route} className='loading-image' width='70px' />
                    <div style={{ minWidth: 0 }}>
                      <h3 style={blockStyle}>{route.Name}</h3>
                      <p style={blockStyle}>{route.Description}</p>
                      <p style={{ ...blockStyle, fontSize: '90%', paddingTop: '3px', opacity: '0.45' }}>{route.Target}</p>
                    </div>
                  </Stack>
                </Link>
              </Box>
            </Grid2>
          ) : (
            <Grid2 xs={6} sm={4} md={3} lg={2} xl={2} xxl={2} key={route.Name}>
              <Box className='app app-hover' style={{ padding: 25, borderRadius: 5, ...appColor, ...appBorder }}>
                <Link to={getFullOrigin(route)} target='_blank' style={{ textDecoration: 'none', ...appColor }}>
                  <Stack direction='column' spacing={2} alignItems='center'>
                    <ServAppIcon container={container} route={route} className='loading-image' width='70px' />
                    <div style={{ minWidth: 0 }}>
                      <h3 style={blockStyle}>{route.Name}</h3>
                    </div>
                  </Stack>
                </Link>
              </Box>
            </Grid2>
          ))
        );
      })}
    </Grid2>
  );
};

export default AppGrid;
