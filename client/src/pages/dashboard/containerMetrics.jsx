import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

// material-ui
import {
    Box,
    Button,
    Grid,
    Stack,
    Typography,
    CircularProgress
} from '@mui/material';

import * as API from '../../api';
import PlotComponent from './components/plot';
import { formatDate } from './components/utils';

const ContainerMetrics = ({containerName}) => {
    const { t } = useTranslation();
    const [slot, setSlot] = useState('latest');

    const [zoom, setZoom] = useState({
        xaxis: {}
    });

    const [coStatus, setCoStatus] = useState(null);
    const [metrics, setMetrics] = useState(null);

    const resetZoom = () => {
        setZoom({
            xaxis: {}
        });
    }

    const metricsKey = {
      CPU: "cosmos.system.docker.cpu." + containerName,
      RAM: "cosmos.system.docker.ram." + containerName,
      NET_RX: "cosmos.system.docker.netRx." + containerName,
      NET_TX: "cosmos.system.docker.netTx." + containerName,
    };

    const refreshMetrics = () => {
        API.metrics.get([
          "cosmos.system.docker.cpu",
          "cosmos.system.docker.ram",
          "cosmos.system.docker.netRx",
          "cosmos.system.docker.netTx",
        ].map(c => c + "." + containerName)).then((res) => {
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
                <Grid item xs={12}  xl={8} sx={{ mb: -2.25 }}>
                    <Typography variant="h4">{containerName} Monitoring</Typography>
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
                    <PlotComponent xAxis={xAxis} zoom={zoom} setZoom={setZoom} slot={slot} title={t('navigation.monitoring.resourcesTitle')} data={[metrics[metricsKey.CPU], metrics[metricsKey.RAM]]}/>
                </Grid>
               
                <Grid item xs={12} xl={8}>
                    <PlotComponent xAxis={xAxis} zoom={zoom} setZoom={setZoom} slot={slot} title={t('global.network')} data={[metrics[metricsKey.NET_TX], metrics[metricsKey.NET_RX]]}/>
                </Grid>
            </Grid>
        </div>}
      </>
    );
};

export default ContainerMetrics;
