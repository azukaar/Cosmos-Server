import { useEffect, useRef, useState } from 'react';
import { Alert, Box, CircularProgress, Grid, Typography } from '@mui/material';
import { useTranslation } from 'react-i18next';

import * as API from '../../api';
import MetricHeaders from './MetricHeaders';
import TableComponent from './components/table';
import { formatDate } from './components/utils';

export const pickMetrics = (metrics, match) =>
  Object.keys(metrics).filter(match).map((key) => metrics[key]);

// Renders nothing until there is data (TableComponent draws an empty card otherwise).
export const MetricTable = ({ data, ...props }) => {
  if (!data || !data.length) return null;
  return <TableComponent data={data} {...props} />;
};

// Same x-axis shape as the main monitoring dashboard.
export const buildXAxis = (slot) => {
  const xAxis = [];

  if (slot === 'latest') {
    for (let i = 0; i < 100; i++) xAxis.unshift(i);
  } else if (slot === 'hourly') {
    for (let i = 0; i < 48; i++) {
      const now = new Date();
      now.setHours(now.getHours() - i);
      now.setMinutes(0);
      now.setSeconds(0);
      xAxis.unshift(formatDate(now, true));
    }
  } else if (slot === 'daily') {
    for (let i = 0; i < 30; i++) {
      const now = new Date();
      now.setDate(now.getDate() - i);
      xAxis.unshift(formatDate(now));
    }
  }

  return xAxis;
};

// Monitoring pane shared by the Constellation feature pages; `children` is a render
// prop handed { metrics, xAxis, slot, zoom, setZoom }. `queries` are metric keys, wildcards allowed.
const FeatureMonitoring = ({ title, queries, note, children }) => {
  const { t } = useTranslation();
  const [slot, setSlot] = useState('latest');
  const [metrics, setMetrics] = useState(null);
  const [zoom, setZoom] = useState({ xaxis: {} });
  // Keep queries in a ref so a fresh array literal from the parent does not re-arm the poll interval.
  const queriesRef = useRef(queries);
  queriesRef.current = queries;

  useEffect(() => {
    let cancelled = false;

    const refresh = () => {
      if (!queriesRef.current || !queriesRef.current.length) {
        setMetrics({});
        return;
      }
      API.metrics.get(queriesRef.current).then((res) => {
        if (cancelled) return;
        const next = {};
        (res.data || []).forEach((metric) => {
          next[metric.Key] = metric;
        });
        setMetrics(next);
      }).catch(() => {
        if (!cancelled) setMetrics({});
      });
    };

    refresh();
    const interval = setInterval(refresh, 10000);

    return () => {
      cancelled = true;
      clearInterval(interval);
    };
  }, []);

  const xAxis = buildXAxis(slot);

  if (!metrics) {
    return <Box style={{ display: 'flex', justifyContent: 'center', marginTop: '150px' }}>
      <CircularProgress size={100} />
    </Box>;
  }

  return <div style={{ zIndex: 2, position: 'relative' }}>
    <Grid container rowSpacing={4.5} columnSpacing={2.75}>
      <Grid item xs={12} sx={{ mb: -2.25 }}>
        <Typography variant="h4">{title}</Typography>
        <MetricHeaders loaded={metrics} slot={slot} setSlot={setSlot} zoom={zoom} setZoom={setZoom} />
      </Grid>

      {note && <Grid item xs={12}>
        <Alert severity="info">{note}</Alert>
      </Grid>}

      {Object.keys(metrics).length === 0 && <Grid item xs={12}>
        <Alert severity="info">{t('navigation.monitoring.noMetrics')}</Alert>
      </Grid>}

      <Grid item xs={12}>
        {children({ metrics, xAxis, slot, zoom, setZoom })}
      </Grid>
    </Grid>
  </div>;
};

export default FeatureMonitoring;
