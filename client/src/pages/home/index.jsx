import { useParams } from "react-router";
import Back from "../../components/back";
import { Alert, Box, CircularProgress, Grid, Stack, useTheme } from "@mui/material";
import { useEffect, useState } from "react";
import * as API  from "../../api";
import wallpaper from '../../assets/images/wallpaper.png';
import Grid2 from "@mui/material/Unstable_Grid2/Grid2";
import { getFaviconURL } from "../../utils/routes";
import { Link } from "react-router-dom";
import { getFullOrigin } from "../../utils/routes";
import IsLoggedIn from "../../isLoggedIn";


const HomeBackground = () => {
    const theme = useTheme();
    return (
        <Box sx={{ position: 'fixed', float: 'left',  overflow: 'hidden',  zIndex: 0, top: 0, left: 0, right: 0, bottom: 0 }}>
            <img src={wallpaper} style={{ display:'inline'}} alt="Cosmos" width="100%" height="100%" />
        </Box>
    );
};

const blockStyle = {
    margin: 0, 
    whiteSpace: 'nowrap',
    overflow: 'hidden',
    textOverflow: 'ellipsis',
    maxWidth: '78%',
    verticalAlign: 'middle',
}

const HomePage = () => {
  const { routeName } = useParams();
  const [config, setConfig] = useState(null);
  const [coStatus, setCoStatus] = useState(null);

  const refreshStatus = () => {
      API.getStatus().then((res) => {
          setCoStatus(res.data);
      });
  }

  const refreshConfig = () => {
    API.config.get().then((res) => {
      setConfig(res.data);
    });
  };
  
  let routes = config && (config.HTTPConfig.ProxyConfig.Routes || []);

  useEffect(() => {
    refreshConfig();
    refreshStatus();
  }, []);
  

  return <Stack spacing={2} >
    <IsLoggedIn />
    <HomeBackground />
    <style>
        {`header {
            background: rgba(0.2,0.2,0.2,0.2) !important;
            border-bottom-color: rgba(0.4,0.4,0.4,0.4) !important;
            color: white !important;
        }

        header .MuiChip-label  {
            color: #eee !important;
        }

        header .MuiButtonBase-root {
            color: #eee !important;
            background: rgba(0.2,0.2,0.2,0.2) !important;
        }

        .app {
            transition: background 0.1s ease-in-out;
            transition: transform 0.1s ease-in-out;
        }
        
        .app:hover {
            cursor: pointer;
            background: rgba(0.4,0.4,0.4,0.4) !important;
            transform: scale(1.05);
        }
    `}
    </style>
    <Stack style={{zIndex: 2}} spacing={1}>
        {coStatus && !coStatus.database && (
            <Alert severity="error">
                No Database is setup for Cosmos! User Management and Authentication will not work.<br />
                You can either setup the database, or disable user management in the configuration panel.<br />
            </Alert>
        )}

        {coStatus && coStatus.letsencrypt && (
            <Alert severity="error">
                You have enabled Let's Encrypt for automatic HTTPS Certificate. You need to provide the configuration with an email address to use for Let's Encrypt in the configs.
            </Alert>
        )}

        {coStatus && coStatus.newVersionAvailable && (
            <Alert severity="warning">
                A new version of Cosmos is available! Please update to the latest version to get the latest features and bug fixes.
            </Alert>
        )}

        {coStatus && coStatus.needsRestart && (
            <Alert severity="warning">
                You have made changes to the configuration that require a restart to take effect. Please restart Cosmos to apply the changes.
            </Alert>
        )}

        {coStatus && coStatus.domain && (
            <Alert severity="error">
                You are using localhost or 0.0.0.0 as a hostname in the configuration. It is recommended that you use a domain name instead.
            </Alert>
        )}

        {coStatus && !coStatus.docker && (
            <Alert severity="error">
                Docker is not connected! Please check your docker connection.<br/>
                Did you forget to add <pre>-v /var/run/docker.sock:/var/run/docker.sock</pre> to your docker run command?<br />
                if your docker daemon is running somewhere else, please add <pre>-e DOCKER_HOST=...</pre> to your docker run command.
            </Alert>
        )}
    </Stack>
    
    <Grid2 container spacing={2} style={{zIndex: 2}}>
        {config && config.HTTPConfig.ProxyConfig.Routes.map((route) => {
            return <Grid2 item xs={12} sm={6} md={4} lg={3} xl={2} key={route.Name}>
                <Box className='app' style={{padding: 10, color: 'white', background: 'rgba(0,0,0,0.35)', borderRadius: 5}}>
                    <Link to={getFullOrigin(route)} target="_blank" style={{textDecoration: 'none', color: 'white'}}>
                        <Stack direction="row" spacing={2} alignItems="center">
                            <img src={getFaviconURL(route)} width="64px" />

                            <div style={{width: '100%'}}>
                                <h3 style={blockStyle}>{route.Name}</h3>
                                <p style={blockStyle}>{route.Description}</p>
                                <p style={blockStyle}>{route.Target}</p>
                            </div>
                        </Stack>
                    </Link>
                </Box>
            </Grid2>
        })}
    </Grid2>
  </Stack>
}

export default HomePage;