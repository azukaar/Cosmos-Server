import { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';

// material-ui
import {
    Avatar,
    AvatarGroup,
    Box,
    Button,
    Grid,
    Typography,
    Alert,
    LinearProgress,
    CircularProgress
} from '@mui/material';

import * as API from '../../api';
import { formatDate } from './components/utils';
import ResourceDashboard from './resourceDashboard';
import PrettyTabbedView from '../../components/tabbedView/tabbedView';
import ProxyDashboard from './proxyDashboard';
import AlertPage from './AlertPage';
import EventsExplorer from './eventsExplorer';
import MetricHeaders from './MetricHeaders';

const DashboardDefault = () => {
    const { t } = useTranslation();
    const [value, setValue] = useState('today');
    const [slot, setSlot] = useState('latest');
    const [currentTab, setCurrentTab] = useState(0);
    const currentTabRef = useRef(currentTab);
    const [coStatus, setCoStatus] = useState(null);
    const [metrics, setMetrics] = useState(null);
    const [isCreatingDB, setIsCreatingDB] = useState(false);

    const [zoom, setZoom] = useState({
        xaxis: {}
    });


    const resetZoom = () => {
        setZoom({
            xaxis: {}
        });
    }

    const refreshMetrics = (override) => {
        let todo = [
            ["cosmos.system.*"],
            ["cosmos.proxy.*"],
            [],
            [],
        ]

        let t = typeof override === 'number' ? override : currentTabRef.current;

        if (t < 2)
            API.metrics.get(todo[t]).then((res) => {
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
    };

    const refreshStatus = () => {
        API.getStatus().then((res) => {
            setCoStatus(res.data);
        });
    }

    useEffect(() => {
        refreshStatus();
        
        let interval = setInterval(() => {
            refreshMetrics();
        }, 10000);

        refreshMetrics();

        return () => {
            clearInterval(interval);
        };
    }, []);
    
    useEffect(() => {
        currentTabRef.current = currentTab;
    }, [currentTab]);

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
                <Grid item xs={12} sx={{ mb: -2.25 }}>
                    <Typography variant="h4">{t('navigation.monitoringTitle')}</Typography>
                    {currentTab <= 2 && <MetricHeaders loaded={metrics} slot={slot} setSlot={setSlot} zoom={zoom} setZoom={setZoom} />}
                    {currentTab > 2 && <div style={{height: 41}}></div>}
                </Grid>


                <Grid item xs={12} md={12} lg={12}>
                    <PrettyTabbedView 
                        currentTab={currentTab}
                        setCurrentTab={(tab) => {
                            setCurrentTab(tab)
                            refreshMetrics(tab);
                        }}
                        fullwidth
                        isLoading={!metrics}
                        tabs={[
                            {
                                title: t('navigation.monitoring.resourcesTitle'),
                                children: <ResourceDashboard xAxis={xAxis} zoom={zoom} setZoom={setZoom} slot={slot} metrics={metrics} />
                            },
                            {
                                title: t('navigation.monitoring.proxyTitle'),
                                children: <ProxyDashboard xAxis={xAxis} zoom={zoom} setZoom={setZoom} slot={slot} metrics={metrics} />
                            },
                            {
                                title: t('navigation.monitoring.eventsTitle'),
                                children: <EventsExplorer xAxis={xAxis} zoom={zoom} setZoom={setZoom} slot={slot} metrics={metrics} />
                            },
                            {
                                title: t('navigation.monitoring.alertsTitle'),
                                children: <AlertPage />
                            },
                        ]}
                    />
                </Grid>
            </Grid>
        </div>}
      </>
    );
};

export default DashboardDefault;
