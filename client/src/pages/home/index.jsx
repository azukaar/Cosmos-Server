import { useParams } from "react-router";
import Back from "../../components/back";
import { Alert, Box, CircularProgress, Grid, Stack, Typography, useMediaQuery, useTheme } from "@mui/material";
import { lazy, useEffect, useState } from "react";
import * as API from "../../api";
import wallpaper from '../../assets/images/wallpaper2.jpg';
import wallpaperLight from '../../assets/images/wallpaper2_light.jpg';
import Grid2 from "@mui/material/Unstable_Grid2/Grid2";
import { getFaviconURL } from "../../utils/routes";
import { Link } from "react-router-dom";
import { getFullOrigin } from "../../utils/routes";
import { ServAppIcon } from "../../utils/servapp-icon";
import Chart from 'react-apexcharts';
import { useClientInfos } from "../../utils/hooks";
import { FormaterForMetric, formatDate } from "../dashboard/components/utils";
import MiniPlotComponent from "../dashboard/components/mini-plot";
import Migrate014 from "./migrate014";
import { Trans, useTranslation } from 'react-i18next';


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
    const { t } = useTranslation();
    const { routeName } = useParams();
    const [servApps, setServApps] = useState([]);
    const [config, setConfig] = useState(null);
    const [coStatus, setCoStatus] = useState(null);
    const [containers, setContainers] = useState(null);
    const theme = useTheme();
    const isDark = theme.palette.mode === 'dark';
    const isMd = useMediaQuery(theme.breakpoints.up('md'));
    const {role} = useClientInfos();
    const isAdmin = role === "2";
    const [metrics, setMetrics] = useState(null);

    const blockStyle = {
        margin: 0,
        whiteSpace: 'nowrap',
        overflow: 'hidden',
        textOverflow: 'ellipsis',
        verticalAlign: 'middle',
    }

    const appColor = isDark ? {
        color: 'white',
        background: 'rgba(10,10,10,0.42)',
    } : {
        color: 'black',
        background: 'rgba(245,245,245,0.42)',
    }


    const appBorder = isDark ? {
    } : {
    }


    const refreshStatus = () => {
        API.getStatus().then((res) => {
            setCoStatus(res.data);
        });
    }

    function getMetrics() {
        if(!isAdmin) return;
        
        API.metrics.get([
            "cosmos.system.cpu.0",
            "cosmos.system.ram",
            "cosmos.system.netTx",
            "cosmos.system.netRx",
            "cosmos.proxy.all.success",
            "cosmos.proxy.all.error",
        ]).then((res) => {
            setMetrics(prevMetrics => {
                let finalMetrics = prevMetrics ? { ...prevMetrics } : {};
                if(res.data) {
                    res.data.forEach((metric) => {
                        finalMetrics[metric.Key] = metric;
                    });
                    
                    return finalMetrics;
                }
            });
        });
    }

    useEffect(() => {
        refreshStatus();
        
        let interval = setInterval(() => {
            if(isMd)
                getMetrics();
        }, 10000);

        if(isMd) getMetrics();

        return () => {
            clearInterval(interval);
        };
    }, []);

    const refreshConfig = () => {
        if(isAdmin) {
            API.docker.list().then((res) => {
                setServApps(res.data);
            });
        } else {
            setServApps([]);
        }
        API.config.get().then((res) => {
            setConfig(res.data);
        });
    };


    let routes = config && (config.HTTPConfig.ProxyConfig.Routes || []);

    useEffect(() => {
        refreshConfig();
        refreshStatus();
    }, []);

    const primCol = theme.palette.primary.main.replace('rgb(', 'rgba(')
    const secCol = theme.palette.secondary.main.replace('rgb(', 'rgba(')
    
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
                            return val + "%"
                        },
                        color: "#111",
                        offsetY: 7,
                        fontSize: "18px",
                        show: true
                    }
                }
            }
        },
        fill: {
            colors: [primCol],
            type: "gradient",
            gradient: {
                shade: "dark",
                type: "horizontal",
                shadeIntensity: 0.5,
                gradientToColors: [secCol], // A light green color as the end of the gradient.
                inverseColors: false,
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
    
    let latestCPU, latestRAM, latestRAMRaw, maxRAM, maxRAMRaw = 0;

    if(isAdmin && metrics) {
    
        if(metrics["cosmos.system.cpu.0"] && metrics["cosmos.system.cpu.0"].Values && metrics["cosmos.system.cpu.0"].Values.length > 0)
            latestCPU = metrics["cosmos.system.cpu.0"].Values[metrics["cosmos.system.cpu.0"].Values.length - 1].Value;
        
        if(metrics["cosmos.system.ram"] && metrics["cosmos.system.ram"].Values && metrics["cosmos.system.ram"].Values.length > 0) {
            let formatRAM = metrics && FormaterForMetric(metrics["cosmos.system.ram"], false);
            latestRAMRaw = metrics["cosmos.system.ram"].Values[metrics["cosmos.system.ram"].Values.length - 1].Value;
            latestRAM = formatRAM(metrics["cosmos.system.ram"].Values[metrics["cosmos.system.ram"].Values.length - 1].Value);
            maxRAM = formatRAM(metrics["cosmos.system.ram"].Max);
            maxRAMRaw = metrics["cosmos.system.ram"].Max;
        }
    }

    return <Stack spacing={2} style={{maxWidth: '1450px', margin:'auto'}}>
        <HomeBackground status={coStatus} />
        <TransparentHeader />

        <Stack style={{ zIndex: 2, padding: '0px 8px'}} spacing={1}>
            {isAdmin && coStatus && !coStatus.database && (
                <Alert severity="error">
                    {t('navigation.home.dbCantConnectError')}
                </Alert>
            )}

            {isAdmin && coStatus && coStatus.letsencrypt && (
                <Alert severity="error">
                    {t('navigation.home.LetsEncryptEmailError')}
                </Alert>
            )}

            {isAdmin && coStatus && (coStatus.backup_status != "") && (
                <Alert severity="error">
                    {coStatus.backup_status}
                </Alert>
            )}

            {isAdmin && coStatus && coStatus.LetsEncryptErrors && coStatus.LetsEncryptErrors.length > 0 && (
                <Alert severity="error">
                    {t('navigation.home.LetsEncryptError')}
                    {coStatus.LetsEncryptErrors.map((err) => {
                        return <div> - {err}</div>
                    })}
                </Alert>
            )}

            {isAdmin && coStatus && coStatus.newVersionAvailable && (
                <Alert severity="warning">
                    {t('navigation.home.newCosmosVersionError')}
                </Alert>
            )}

            {isAdmin && coStatus && !coStatus.hostmode && config && (
                <Alert severity="warning">
                    {t('navigation.home.cosmosNotDockerHostError')} <br />
                    <Migrate014 config={config} />
                </Alert>
            )}

            {isAdmin && coStatus && coStatus.needsRestart && (
                <Alert severity="warning">
                    {t('navigation.home.configChangeRequiresRestartError')}
                </Alert>
            )}

            {isAdmin && coStatus && coStatus.domain && (
                <Alert severity="error">
                    {t('navigation.home.localhostnotRecommendedError')}
                </Alert>
            )}

            {isAdmin && coStatus && !coStatus.docker && (
                <Alert severity="error"><Trans i18nKey="newInstall.dockerNotConnected" /></Alert>
            )}
        </Stack>

        <Grid2 container spacing={2} style={{ zIndex: 2 }}>
            {isAdmin && coStatus && !coStatus.MonitoringDisabled && (<>
                {isMd && !metrics && (<>
                    <Grid2 item xs={12} sm={6} md={6} lg={3} xl={3} xxl={3} key={'000'}>
                        <Box className='app' style={{height: '106px', borderRadius: 5, ...appColor }}>
                            <Stack direction="row" justifyContent={'space-between'} alignItems={'center'} style={{ height: "100%" }}>
                                <Stack style={{paddingLeft: '20px'}} spacing={0}>
                                <div style={{fontSize: '18px', fontWeight: "bold"}}>{t('global.CPU')}</div>
                                <div>-</div>
                                <div>-</div>
                                </Stack>
                                <div style={{height: '97px'}}>
                                    -
                                </div>
                            </Stack>
                        </Box>
                    </Grid2>
                    <Grid2 item xs={12} sm={6} md={6} lg={3} xl={3} xxl={3} key={'001'}>
                        <Box className='app' style={{height: '106px', borderRadius: 5, ...appColor }}>
                            <Stack direction="row" justifyContent={'space-between'} alignItems={'center'} style={{ height: "100%" }}>
                                <Stack style={{paddingLeft: '20px'}} spacing={0}>
                                    <div style={{fontSize: '18px', fontWeight: "bold"}}>{t('global.RAM')}</div>
                                    <div>{t('navigation.home.availRam')}: -</div>
                                    <div>{t('navigation.home.usedRam')}: -</div>
                                </Stack>
                                <div style={{height: '97px'}}>
                                    -
                                </div>
                            </Stack>
                        </Box>
                    </Grid2>
                    <Grid2 item xs={12} sm={6} md={6} lg={3} xl={3} xxl={3} key={'001'}>
                        <Box className='app' style={{height: '106px',borderRadius: 5, ...appColor }}>
                        <Stack direction="row" justifyContent={'center'} alignItems={'center'} style={{ height: "100%" }}>
                            -
                        </Stack>
                        </Box>
                    </Grid2>
                    <Grid2 item xs={12} sm={6} md={6} lg={3} xl={3} xxl={3} key={'001'}>
                        <Box className='app' style={{height: '106px',borderRadius: 5, ...appColor }}>
                        <Stack direction="row" justifyContent={'center'} alignItems={'center'} style={{ height: "100%" }}>
                            -
                        </Stack>
                        </Box>
                    </Grid2>
                </>)}

                {isMd && metrics && (<>
                    <Grid2 item xs={12} sm={6} md={6} lg={3} xl={3} xxl={3} key={'000'}>
                        <Box className='app' style={{height: '106px', borderRadius: 5, ...appColor }}>
                            <Stack direction="row" justifyContent={'space-between'} alignItems={'center'} style={{ height: "100%" }}>
                                <Stack style={{paddingLeft: '20px'}} spacing={0}>
                                <div style={{fontSize: '18px', fontWeight: "bold"}}>{t('global.CPU')}</div>
                                <div>{coStatus.CPU}</div>
                                <div>{coStatus.AVX ? t('navigation.home.Avx') : t('navigation.home.noAvx')}</div>
                                </Stack>
                                <div style={{height: '97px'}}>
                                    <Chart
                                        options={optionsRadial}
                                        // series={[parseInt(
                                        //     coStatus.resources.ram / (coStatus.resources.ram + coStatus.resources.ramFree) * 100
                                        // )]}
                                        series={[latestCPU]}
                                        type="radialBar"
                                        height={120}
                                        width={120}
                                    />
                                </div>
                            </Stack>
                        </Box>
                    </Grid2>
                    <Grid2 item xs={12} sm={6} md={6} lg={3} xl={3} xxl={3} key={'001'}>
                        <Box className='app' style={{height: '106px', borderRadius: 5, ...appColor }}>
                            <Stack direction="row" justifyContent={'space-between'} alignItems={'center'} style={{ height: "100%" }}>
                                <Stack style={{paddingLeft: '20px'}} spacing={0}>
                                    <div style={{fontSize: '18px', fontWeight: "bold"}}>{t('global.RAM')}</div>
                                    <div>{t('navigation.home.availRam')}: <strong>{maxRAM}</strong></div>
                                    <div>{t('navigation.home.usedRam')}: <strong>{latestRAM}</strong></div>
                                </Stack>
                                <div style={{height: '97px'}}>
                                    <Chart
                                        options={optionsRadial}
                                        // series={[parseInt(
                                        //     coStatus.resources.ram / (coStatus.resources.ram + coStatus.resources.ramFree) * 100
                                        // )]}
                                        series={[parseInt(latestRAMRaw / maxRAMRaw * 100)]}
                                        type="radialBar"
                                        height={120}
                                        width={120}
                                    />
                                </div>
                            </Stack>
                        </Box>
                    </Grid2>
                    <Grid2 item xs={12} sm={6} md={6} lg={3} xl={3} xxl={3} key={'001'}>
                        <Box className='app' style={{height: '106px',borderRadius: 5, ...appColor }}>
                        <Stack direction="row" justifyContent={'center'} alignItems={'center'} style={{ height: "100%" }}>
                            <MiniPlotComponent noBackground title={t('navigation.home.network')} agglo metrics={[
                                "cosmos.system.netTx",
                                "cosmos.system.netRx",
                            ]} labels={{
                                ["cosmos.system.netTx"]: t('navigation.home.trsNet')+":", 
                                ["cosmos.system.netRx"]: t('navigation.home.rcvNet')+":"
                            }}/>
                        </Stack>
                        </Box>
                    </Grid2>
                    <Grid2 item xs={12} sm={6} md={6} lg={3} xl={3} xxl={3} key={'001'}>
                        <Box className='app' style={{height: '106px',borderRadius: 5, ...appColor }}>
                        <Stack direction="row" justifyContent={'center'} alignItems={'center'} style={{ height: "100%" }}>
                            <MiniPlotComponent noBackground title='URLS' agglo metrics={[
                                "cosmos.proxy.all.success",
                                "cosmos.proxy.all.error",
                            ]} labels={{
                                ["cosmos.proxy.all.success"]: "ok:", 
                                ["cosmos.proxy.all.error"]: "err:"
                            }}/>
                        </Stack>
                        </Box>
                    </Grid2>
                
                </>)}
            </>)}
            
            {config && servApps && routes.map((route) => {
                let skip = route.Mode == "REDIRECT";
                let containerName;
                let container;
                if (route.Mode == "SERVAPP") {
                    containerName = route.Target.split(':')[1].slice(2);
                    container = servApps.find((c) => c.Names.includes('/' + containerName));
                    // TOOD: rework, as it prevents users from seeing the apps
                    // if (!container || container.State != "running") {
                    //     skip = true
                    // }
                }
                
                if (route.HideFromDashboard) 
                    skip = true;

                return !skip && coStatus && (coStatus.homepage.Expanded ?
                
                <Grid2 item xs={12} sm={6} md={4} lg={3} xl={3} xxl={3} key={route.Name}>
                    <Box className='app app-hover' style={{ padding: 25, borderRadius: 5, ...appColor, ...appBorder }}>
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
                :
                <Grid2 item xs={6} sm={4} md={3} lg={2} xl={2} xxl={2} key={route.Name}>
                    <Box className='app app-hover' style={{ padding: 25, borderRadius: 5, ...appColor, ...appBorder }}>
                        <Link to={getFullOrigin(route)} target="_blank" style={{ textDecoration: 'none', ...appColor }}>
                            <Stack direction="column" spacing={2} alignItems="center">
                                <ServAppIcon container={container} route={route} className="loading-image" width="70px" />
                                <div style={{ minWidth: 0 }}>
                                    <h3 style={blockStyle}>{route.Name}</h3>
                                </div>
                            </Stack>
                        </Link>
                    </Box>
                </Grid2>)
            })}

            {config && routes.length === 0 && (
                <Grid2 item xs={12} sm={12} md={12} lg={12} xl={12}>
                    <Box style={{ padding: 10, borderRadius: 5, ...appColor }}>
                        <Stack direction="row" spacing={2} alignItems="center">
                            <div style={{ minWidth: 0 }}>
                                <h3 style={blockStyle}>{t('navigation.home.noAppsTitle')}</h3>
                                <p style={blockStyle}>{t('navigation.home.noApps')}</p>
                            </div>
                        </Stack>
                    </Box>
                </Grid2>
            )}
        </Grid2>
    </Stack>
}

export default HomePage;