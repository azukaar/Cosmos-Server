import { useParams } from "react-router";
import Back from "../../components/back";
import { Alert, Box, CircularProgress, Grid, Stack, useTheme } from "@mui/material";
import { useEffect, useState } from "react";
import * as API from "../../api";
import wallpaper from '../../assets/images/wallpaper2.jpg';
import wallpaperLight from '../../assets/images/wallpaper2_light.jpg';
import Grid2 from "@mui/material/Unstable_Grid2/Grid2";
import { getFaviconURL } from "../../utils/routes";
import { Link } from "react-router-dom";
import { getFullOrigin } from "../../utils/routes";
import IsLoggedIn from "../../isLoggedIn";


const HomeBackground = () => {
    const theme = useTheme();
    const isDark = theme.palette.mode === 'dark';
    return (
        <Box sx={{ position: 'fixed', float: 'left', overflow: 'hidden', zIndex: 0, top: 0, left: 0, right: 0, bottom: 0,
            // gradient
            // backgroundImage: isDark ? 
            //     `linear-gradient(#371d53, #26143a)` : 
            //     `linear-gradient(#e6d3fb, #c8b0e2)`,
        }}>
            <img src={isDark ? wallpaper : wallpaperLight } style={{ display: 'inline' }} alt="Cosmos" width="100%" height="100%" />
        </Box>
    );
};

const HomePage = () => {
    const { routeName } = useParams();
    const [serveApps, setServeApps] = useState([]);
    const [config, setConfig] = useState(null);
    const [coStatus, setCoStatus] = useState(null);
    const [containers, setContainers] = useState(null);
    const theme = useTheme();
    const isDark = theme.palette.mode === 'dark';

    const blockStyle = {
        margin: 0,
        whiteSpace: 'nowrap',
        overflow: 'hidden',
        textOverflow: 'ellipsis',
        maxWidth: '78%',
        verticalAlign: 'middle',
    }

    const appColor =  isDark ? {
        color: 'white', 
        background: 'rgba(0,0,0,0.35)',
    } : {
        color: 'black',
        background: 'rgba(255,255,255,0.35)',
    }

    const backColor =  isDark ? '0,0,0' : '255,255,255';
    const textColor =  isDark ? 'white' : 'dark';


    const refreshStatus = () => {
        API.getStatus().then((res) => {
            setCoStatus(res.data);
        });
    }

    const refreshConfig = () => {
        API.docker.list().then((res) => {
            setServeApps(res.data);
        });
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
            background: rgba(${backColor},0.35) !important;
            border-bottom-color: rgba(${backColor},0.4) !important;
            color: ${textColor} !important;
            font-weight: bold;
        }

        header .MuiChip-label  {
            color: ${textColor} !important;
        }

        header .MuiButtonBase-root, header .MuiChip-colorDefault  {
            color: ${textColor} !important;
            background: rgba(${backColor},0.5) !important;
        }

        .app {
            transition: background 0.1s ease-in-out;
            transition: transform 0.1s ease-in-out;
        }
        
        .app:hover {
            cursor: pointer;
            background: rgba(${backColor},0.8) !important;
            transform: scale(1.05);
        }

        .MuiAlert-standard {
            background: rgba(${backColor},0.35) !important;
            color: ${textColor} !important;
            font-weight: bold;
        }
    `}
        </style>
        <Stack style={{ zIndex: 2 }} spacing={1}>
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
                    Docker is not connected! Please check your docker connection.<br />
                    Did you forget to add <pre>-v /var/run/docker.sock:/var/run/docker.sock</pre> to your docker run command?<br />
                    if your docker daemon is running somewhere else, please add <pre>-e DOCKER_HOST=...</pre> to your docker run command.
                </Alert>
            )}
        </Stack>

        <Grid2 container spacing={2} style={{ zIndex: 2 }}>
            {config && serveApps && config.HTTPConfig.ProxyConfig.Routes.map((route) => {
                let skip = false;
                if(route.Mode == "SERVAPP") {
                    const containerName = route.Target.split(':')[1].slice(2);
                    console.log('try ' + containerName)
                    const container = serveApps.find((c) => c.Names.includes('/' + containerName));
                    console.log('found ' + container)
                    if(!container || container.State != "running") {
                        skip = true
                    }
                }
                return !skip && <Grid2 item xs={12} sm={6} md={4} lg={3} xl={3} xxl={2} key={route.Name}>
                    <Box className='app' style={{ padding: 10,  borderRadius: 5, ...appColor }}>
                        <Link to={getFullOrigin(route)} target="_blank" style={{ textDecoration: 'none', ...appColor }}>
                            <Stack direction="row" spacing={2} alignItems="center">
                                <img className="loading-image" alt="" src={getFaviconURL(route)} width="64px" height="64px"/>

                                <div style={{minWidth: 0 }}>
                                    <h3 style={blockStyle}>{route.Name}</h3>
                                    <p style={blockStyle}>{route.Description}</p>
                                    <p style={blockStyle}>{route.Target}</p>
                                </div>
                            </Stack>
                        </Link>
                    </Box>
                </Grid2>
            })}

            {config && config.HTTPConfig.ProxyConfig.Routes.length === 0 && (
                <Grid2 item xs={12} sm={12} md={12} lg={12} xl={12}>
                    <Box style={{ padding: 10, borderRadius: 5, ...appColor }}>
                        <Stack direction="row" spacing={2} alignItems="center">
                            <div style={{ width: '100%' }}>
                                <h3 style={blockStyle}>No Apps</h3>
                                <p style={blockStyle}>You have no apps configured. Please add some apps in the configuration panel.</p>
                            </div>
                        </Stack>
                    </Box>
                </Grid2>
            )}
        </Grid2>
    </Stack>
}

export default HomePage;