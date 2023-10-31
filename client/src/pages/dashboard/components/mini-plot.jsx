import React, { useEffect, useState, useMemo  } from 'react';
import { useInView } from 'react-intersection-observer';

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
  Alert
} from '@mui/material';
import MainCard from '../../../components/MainCard';

// material-ui
import { useTheme } from '@mui/material/styles';

// third-party
import ReactApexChart from 'react-apexcharts';
import { FormaterForMetric, formatDate, toUTC } from './utils';

import * as API from '../../../api';

const _MiniPlotComponent = ({metrics, labels}) => {
  const theme = useTheme();
  const { primary, secondary } = theme.palette.text;
  const [dataMetrics, setDataMetrics] = useState([]);
  const [series, setSeries] = useState([]);
  const slot = 'hourly';
  const [ref, inView] = useInView();

  useEffect(() => {
    if(!inView) return;
    
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

    let xAxisLength = xAxis.length;
    xAxis = xAxis.filter((x, i) => i >= xAxisLength / 4);

    // request data
    API.metrics.get(metrics).then((res) => {
      const toProcess = res.data || [];
      setDataMetrics(toProcess);

      const _series = toProcess.map((serie) => ({
        name: 'Value',
        data:  xAxis.map((date) => {
          if(slot === 'latest') {
            return serie.Values[serie.Values.length - 1 - date] ? 
              serie.Values[serie.Values.length - 1 - date].Value :
              0;
          } else {
            let key = slot === 'hourly' ? "hour_" : "day_";
            let k = key + toUTC(date, slot === 'hourly');
            if (k in serie.ValuesAggl) {
              return serie.ValuesAggl[k].Value;
            } else {
              return 0;
            }
          }
        })
      }));

      setSeries(_series);
    });
  }, [metrics, inView]);

  const chartOptions = {
    colors: [
      theme.palette.primary.main.replace('rgb(', 'rgba('), 
      theme.palette.secondary.main.replace('rgb(', 'rgba(')
    ],
    chart: {
      type: 'area',
      toolbar: {
        show: false,
      },
      zoom: {
        enabled: false
      },
    },
    grid: {
      show: false,
    },
    dataLabels: {
      enabled: false
    },
    stroke: {
      show: true,
      width: 2,
      curve: 'smooth'
    },
    xaxis: {
      labels: {
        show: false
      },
      axisBorder: {
        show: false
      },
      axisTicks: {
        show: false
      }
    },
    yaxis: dataMetrics.map((data) => ({
      labels: {
        show: false
      },
      axisBorder: {
        show: false
      },
      axisTicks: {
        show: false
      }
    })),
    tooltip: {
      enabled: false
    },
    legend: {
      show: false
    },
    markers: {
      size: 0
    },
    responsive: [{
      breakpoint: undefined,
      options: {},
    }]
  };

  const formaters = dataMetrics.map((data) => FormaterForMetric(data));

  return <Stack direction='row' spacing={3} 
                alignItems='center' sx={{padding: '0px 20px', width: '100%', backgroundColor: 'rgba(0,0,0,0.1)'}}
                justifyContent={'space-around'}>

        {dataMetrics && dataMetrics.map((dataMetric, di) => 
        <Stack direction='column' justifyContent={'center'} alignItems={'center'} spacing={0} style={{
          width: '60px',
        }}>
          <div style={{
              fontWeight: 'bold',
              fontSize: '110%',
              whiteSpace: 'nowrap',
            }}>{formaters[di](dataMetric.Values[dataMetric.Values.length - 1].Value)}</div>
          <div>{(labels && labels[dataMetric.Key]) || dataMetric.Label}</div>
        </Stack>)}

      <div style={{
        margin: '-10px 0px -20px 10px',
        flexGrow: 1,
      }} ref={ref}>
        <ReactApexChart options={chartOptions} series={series} type="line" height={90} />
      </div>
    </Stack>
}

const MiniPlotComponent = ({ metrics, labels }) => {
  const memoizedComponent = useMemo(() => <_MiniPlotComponent metrics={metrics} labels={labels} />, [metrics]);
  return memoizedComponent;
};

export default MiniPlotComponent;