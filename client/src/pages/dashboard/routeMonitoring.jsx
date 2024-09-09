import { useEffect, useState } from 'react';

// material-ui
import {
    Avatar,
    AvatarGroup,
    Box,
    Button,
    Grid,
    List,
    ListItemAvatar,
    ListItemButton,
    ListItemSecondaryAction,
    ListItemText,
    MenuItem,
    Stack,
    TextField,
    Typography,
    Alert,
    LinearProgress,
    CircularProgress
} from '@mui/material';

import * as API from '../../api';
import AnimateButton from '../../components/@extended/AnimateButton';
import PlotComponent from './components/plot';
import TableComponent from './components/table';
import { HomeBackground, TransparentHeader } from '../home';
import { formatDate } from './components/utils';
import { useTranslation } from 'react-i18next';

// avatar style
const avatarSX = {
    width: 36,
    height: 36,
    fontSize: '1rem'
};

// action style
const actionSX = {
    mt: 0.75,
    ml: 1,
    top: 'auto',
    right: 'auto',
    alignSelf: 'flex-start',
    transform: 'none'
};

// sales report status
const status = [
    {
        value: 'today',
        label: 'Today'
    },
    {
        value: 'month',
        label: 'This Month'
    },
    {
        value: 'year',
        label: 'This Year'
    }
];

// ==============================|| DASHBOARD - DEFAULT ||============================== //

const RouteMetrics = ({routeName}) => {
    const { t } = useTranslation();
    const [value, setValue] = useState('today');
    const [slot, setSlot] = useState('latest');

    const [zoom, setZoom] = useState({
        xaxis: {}
    });

    const [coStatus, setCoStatus] = useState(null);
    const [metrics, setMetrics] = useState(null);
    const [isCreatingDB, setIsCreatingDB] = useState(false);

    const resetZoom = () => {
        setZoom({
            xaxis: {}
        });
    }

    const metricsKey = {
      SUCCESS: "cosmos.proxy.route.success." + routeName,
      ERROR: "cosmos.proxy.route.error." + routeName,
      TIME: "cosmos.proxy.route.time." + routeName,
      BYTES: "cosmos.proxy.route.bytes." + routeName,
    };

    const refreshMetrics = () => {
        API.metrics.get([
          "cosmos.proxy.route.success",
          "cosmos.proxy.route.error",
          "cosmos.proxy.route.time",
          "cosmos.proxy.route.bytes",
        ].map(c => c + "." + routeName)).then((res) => {
            let finalMetrics = {};
            if(res.data) {
                res.data.forEach((metric) => {
                    finalMetrics[metric.Key] = metric;
                });
                setMetrics(finalMetrics);
            }
            setTimeout(refreshMetrics, 10000);
        });
    };

    const refreshStatus = () => {
        API.getStatus().then((res) => {
            setCoStatus(res.data);
        });
    }

    useEffect(() => {
        refreshStatus();
        refreshMetrics();
    }, []);
    
    let xAxis = [];

    if(slot === 'latest') {
      for(let i = 0; i < 100; i++) {
        xAxis.unshift(i);
      }
    }
    else if(slot === 'hourly') {
      for(let i = 0; i < 48; i++) {
        let now = new Date();
        now.setHours(now.getHours() - i);
        now.setMinutes(0);
        now.setSeconds(0);
        xAxis.unshift(formatDate(now, true));
      }
    } else if(slot === 'daily') {
      for(let i = 0; i < 30; i++) {
        let now = new Date();
        now.setDate(now.getDate() - i);
        xAxis.unshift(formatDate(now));
      }
    }

    return (
        <>
        {!metrics && <Box style={{
          width: '100%',
          height: '100%',
          zIndex: 1000,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          marginTop: '150px',
        }}>
          <CircularProgress
            size={100}
          />
        </Box>}
        {metrics && <div style={{zIndex:2, position: 'relative'}}>
            <Grid container rowSpacing={4.5} columnSpacing={2.75} >
                <Grid item xs={12} xl={8} sx={{ mb: -2.25 }}>
                    <Typography variant="h4">{routeName} Monitoring</Typography>
                    <Stack direction="row" alignItems="center" spacing={0} style={{marginTop: 10}}>
                        <Button
                            size="small"
                            onClick={() => {setSlot('latest'); resetZoom()}}
                            color={slot === 'latest' ? 'primary' : 'secondary'}
                            variant={slot === 'latest' ? 'outlined' : 'text'}
                        >
                            {t('navigation.monitoring.latest')}
                        </Button>
                        <Button
                            size="small"
                            onClick={() => {setSlot('hourly'); resetZoom()}}
                            color={slot === 'hourly' ? 'primary' : 'secondary'}
                            variant={slot === 'hourly' ? 'outlined' : 'text'}
                        >
                            {t('navigation.monitoring.hourly')}
                        </Button>
                        <Button
                            size="small"
                            onClick={() => {setSlot('daily'); resetZoom()}}
                            color={slot === 'daily' ? 'primary' : 'secondary'}
                            variant={slot === 'daily' ? 'outlined' : 'text'}
                        >
                            {t('navigation.monitoring.daily')}
                        </Button>

                        {zoom.xaxis.min && <Button
                            size="small"
                            onClick={() => {
                                setZoom({
                                    xaxis: {}
                                });
                            }}
                            color={'primary'}
                            variant={'outlined'}
                        >
                            {t('global.resetZoomButton')}
                        </Button>}
                    </Stack>
                </Grid>
                
                <Grid item xs={12} xl={8}>
                    <PlotComponent xAxis={xAxis} zoom={zoom} setZoom={setZoom} slot={slot} title={'Requests'} data={[metrics[metricsKey.SUCCESS], metrics[metricsKey.ERROR]]}/>
                </Grid>
               
                <Grid item xs={12} xl={8}>
                    <PlotComponent xAxis={xAxis} zoom={zoom} setZoom={setZoom} slot={slot} title={'Resources'} data={[metrics[metricsKey.TIME], metrics[metricsKey.BYTES]]}/>
                </Grid>
            </Grid>
        </div>}
      </>
    );
};

export default RouteMetrics;
