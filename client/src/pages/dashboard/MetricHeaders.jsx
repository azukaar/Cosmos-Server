import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

// material-ui
import {
    Button,
    Stack,
} from '@mui/material';

import { formatDate } from './components/utils';

const MetricHeaders = ({loaded, slot, setSlot, zoom, setZoom}) => {
    const { t } = useTranslation();
    const resetZoom = () => {
        setZoom({
            xaxis: {}
        });
    }

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
        {loaded && <div style={{zIndex:2, position: 'relative'}}>
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
        </div>}
      </>
    );
};

export default MetricHeaders;
