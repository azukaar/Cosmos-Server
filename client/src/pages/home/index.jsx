import { useParams } from "react-router";
import Back from "../../components/back";
import { Alert, Box, CircularProgress, Grid, Stack, Typography, useMediaQuery, useTheme } from "@mui/material";
import { useEffect, useState } from "react";
import * as API from "../../api";
import wallpaper from '../../assets/images/wallpaper2.jpg';
import wallpaperLight from '../../assets/images/wallpaper2_light.jpg';
import Grid2 from "@mui/material/Unstable_Grid2/Grid2";
import { getFaviconURL } from "../../utils/routes";
import { Link } from "react-router-dom";
import { getFullOrigin } from "../../utils/routes";
import IsLoggedIn from "../../isLoggedIn";
import { ServAppIcon } from "../../utils/servapp-icon";
import Chart from 'react-apexcharts';


export const HomeBackground = () => {
    const theme = useTheme();
    const isDark = theme.palette.mode === 'dark';
    const customPaper = API.HOME_BACKGROUND;
    return (
        <Box sx={{
            position: 'fixed', float: 'left', overflow: 'hidden', zIndex: 0, top: 0, left: 0, right: 0, bottom: 0,
            backgroundSize: 'cover',
            backgroundPosition: 'center',
            backgroundRepeat: 'no-repeat',
            backgroundAttachment: 'fixed',
            backgroundImage: customPaper ? `url(${customPaper})` : (isDark ? `url(${wallpaper})` : `url(${wallpaperLight})`),
        }}>
        </Box>
    );
};

export const TransparentHeader = () => {
    const theme = useTheme();
    const isDark = theme.palette.mode === 'dark';

    const backColor = isDark ? '0,0,0' : '255,255,255';
    const textColor = isDark ? 'white' : 'dark';

    return <style>
        {`header {
        background: rgba(${backColor}, 0.40) !important;
        border-bottom-color: rgba(${backColor},0.45) !important;
        color: ${textColor} !important;
        font-weight: bold;
        backdrop-filter: blur(15px);
    }

    header .MuiChip-label  {
        color: ${textColor} !important;
    }

    header .MuiButtonBase-root, header .MuiChip-colorDefault  {
        color: ${textColor} !important;
        background: rgba(${backColor},0.5) !important;
    }

    .app {
        backdrop-filter: blur(15px);
        transition: background 0.1s ease-in-out;
        transition: transform 0.1s ease-in-out;
    }
    
    .app-hover:hover {
        cursor: pointer;
        background: rgba(${backColor},0.8) !important;
        transform: scale(1.05);
    }

    .MuiAlert-standard {
        backdrop-filter: blur(15px);
        background: rgba(${backColor},0.40) !important;
        color: ${textColor} !important;
        font-weight: bold;
    }
`}
    </style>;
}

const HomePage = () => {
    const { routeName } = useParams();
    const [servApps, setServApps] = useState([]);
    const [config, setConfig] = useState(null);
    const [coStatus, setCoStatus] = useState(null);
    const [containers, setContainers] = useState(null);
    const theme = useTheme();
    const isDark = theme.palette.mode === 'dark';
    const isMd = useMediaQuery(theme.breakpoints.up('md'));

    const blockStyle = {
        margin: 0,
        whiteSpace: 'nowrap',
        overflow: 'hidden',
        textOverflow: 'ellipsis',
        verticalAlign: 'middle',
    }

    const appColor = isDark ? {
        color: 'white',
        background: 'rgba(0,0,0,0.40)',
    } : {
        color: 'black',
        background: 'rgba(255,255,255,0.40)',
    }


    const refreshStatus = () => {
        API.getStatus().then((res) => {
            setCoStatus(res.data);
        });
    }

    const refreshConfig = () => {
        API.docker.list().then((res) => {
            setServApps(res.data);
        });
        API.config.get().then((res) => {
            setConfig(res.data);
        });
    };


    let routes = config && (config.HTTPConfig.ProxyConfig.Routes || []);

    useEffect(() => {
        refreshConfig();
        refreshStatus();

        // const interval = setInterval(() => {
        //     refreshStatus();
        // }, 5000);

        // return () => clearInterval(interval);
    }, []);

    const optionsRadial = {
        plotOptions: {
            radialBar: {
                startAngle: -135,
                endAngle: 225,
                hollow: {
                    margin: 0,
                    size: "70%",
                    background: "#fff",
                    image: undefined,
                    imageOffsetX: 0,
                    imageOffsetY: 0,
                    position: "front",
                    dropShadow: {
                        enabled: true,
                        top: 3,
                        left: 0,
                        blur: 4,
                        opacity: 0.24
                    }
                },

                track: {
                    background: "#fff",
                    strokeWidth: "67%",
                    margin: 0, // margin is in pixels
                    dropShadow: {
                        enabled: true,
                        top: -3,
                        left: 0,
                        blur: 4,
                        opacity: 0.35
                    }
                },

                dataLabels: {
                    showOn: "always",
                    name: {
                        show: false,
                        color: "#888",
                        fontSize: "13px"
                    },
                    value: {
                        formatter: function (val) {
                            return val + "%";
                        },
                        color: "#111",
                        offsetY: 9,
                        fontSize: "24px",
                        show: true
                    }
                }
            }
        },
        fill: {
            type: "gradient",
            gradient: {
                shade: "dark",
                type: "horizontal",
                shadeIntensity: 0.5,
                gradientToColors: ["#ABE5A1"],
                inverseColors: true,
                opacityFrom: 1,
                opacityTo: 1,
                stops: [0, 100]
            }
        },
        stroke: {
            lineCap: "round"
        },
        labels: []
    };


    return <Stack spacing={2} >
        <IsLoggedIn />
        <HomeBackground status={coStatus} />
        <TransparentHeader />
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
            {/* {(!isMd || !coStatus || !coStatus.resources || !coStatus.resources.cpu ) && (<>
                <Grid2 item xs={12} sm={6} md={4} lg={3} xl={3} xxl={3} key={'000'}>
                    <Box className='app' style={{ padding: '20px', height: "107px", borderRadius: 5, ...appColor }}>
                        <Stack style={{flexGrow: 1}} spacing={0}>
                            <div style={{fontSize: '18px', fontWeight: "bold"}}>CPU</div>
                            <div>--</div>
                            <div>--</div>
                        </Stack>
                    </Box>
                </Grid2>
                <Grid2 item xs={12} sm={6} md={4} lg={3} xl={3} xxl={3} key={'001'}>
                <Box className='app' style={{  padding: '20px', height: "107px", borderRadius: 5, ...appColor }}>
                        <Stack style={{flexGrow: 1}} spacing={0}>
                        <div style={{fontSize: '18px', fontWeight: "bold"}}>RAM</div>
                            <div>Available: --GB</div>
                            <div>Used: --GB</div>
                        </Stack>
                    </Box>
                </Grid2>
            </>)}

            {isMd && coStatus && coStatus.resources.cpu && (<>
                <Grid2 item xs={12} sm={6} md={4} lg={3} xl={3} xxl={3} key={'000'}>
                    <Box className='app' style={{height: '106px', borderRadius: 5, ...appColor }}>
                    <Stack direction="row" justifyContent={'center'} alignItems={'center'} style={{ height: "100%" }}>
                        <div style={{width: '110px', height: '97px'}}>
                            <Chart
                                options={optionsRadial}
                                series={[parseInt(coStatus.resources.cpu)]}
                                type="radialBar"
                                height={120}
                                width={120}
                            />
                        </div>
                        <Stack style={{flexGrow: 1}} spacing={0}>
                            <div style={{fontSize: '18px', fontWeight: "bold"}}>CPU</div>
                            <div>{coStatus.CPU}</div>
                            <div>{coStatus.AVX ? "AVX Supported" : "No AVX Support"}</div>
                        </Stack>
                    </Stack>
                    </Box>
                </Grid2>
                <Grid2 item xs={12} sm={6} md={4} lg={3} xl={3} xxl={3} key={'001'}>
                    <Box className='app' style={{height: '106px',borderRadius: 5, ...appColor }}>
                    <Stack direction="row" justifyContent={'center'} alignItems={'center'} style={{ height: "100%" }}>
                        <div style={{width: '110px', height: '97px'}}>
                            <Chart
                                options={optionsRadial}
                                series={[parseInt(
                                    coStatus.resources.ram / (coStatus.resources.ram + coStatus.resources.ramFree) * 100
                                )]}
                                type="radialBar"
                                height={120}
                                width={120}
                            />
                        </div>
                        <Stack style={{flexGrow: 1}} spacing={0}>
                            <div style={{fontSize: '18px', fontWeight: "bold"}}>RAM</div>
                            <div>Available: {Math.ceil((coStatus.resources.ram + coStatus.resources.ramFree) / 1024 / 1024 / 1024)}GB</div>
                            <div>Used: {Math.ceil(coStatus.resources.ram / 1024 / 1024 / 1024)}GB</div>
                        </Stack>
                    </Stack>
                    </Box>
                </Grid2>
            
            </>)} */}
            
            {config && servApps && routes.map((route) => {
                let skip = route.Mode == "REDIRECT";
                let containerName;
                let container;
                if (route.Mode == "SERVAPP") {
                    containerName = route.Target.split(':')[1].slice(2);
                    container = servApps.find((c) => c.Names.includes('/' + containerName));
                    if (!container || container.State != "running") {
                        skip = true
                    }
                }
                return !skip && <Grid2 item xs={12} sm={6} md={4} lg={3} xl={3} xxl={3} key={route.Name}>
                    <Box className='app app-hover' style={{ padding: 18, borderRadius: 5, ...appColor }}>
                        <Link to={getFullOrigin(route)} target="_blank" style={{ textDecoration: 'none', ...appColor }}>
                            <Stack direction="row" spacing={2} alignItems="center">
                                <ServAppIcon container={container} route={route} className="loading-image" width="70px" />
                                <div style={{ minWidth: 0 }}>
                                    <h3 style={blockStyle}>{route.Name}</h3>
                                    <p style={blockStyle}>{route.Description}</p>
                                    <p style={{ ...blockStyle, fontSize: '90%', paddingTop: '3px', opacity: '0.45' }}>{route.Target}</p>
                                </div>
                            </Stack>
                        </Link>
                    </Box>
                </Grid2>
            })}

            {config && routes.length === 0 && (
                <Grid2 item xs={12} sm={12} md={12} lg={12} xl={12}>
                    <Box style={{ padding: 10, borderRadius: 5, ...appColor }}>
                        <Stack direction="row" spacing={2} alignItems="center">
                            <div style={{ minWidth: 0 }}>
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