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
  Alert
} from '@mui/material';
import MainCard from '../../../components/MainCard';

// material-ui
import { useTheme } from '@mui/material/styles';

// third-party
import ReactApexChart from 'react-apexcharts';

function formatDate(now, time) {
  // use as UTC
  // now.setMinutes(now.getMinutes() - now.getTimezoneOffset());

  const year = now.getFullYear();
  const month = String(now.getMonth() + 1).padStart(2, '0'); // Months are 0-based
  const day = String(now.getDate()).padStart(2, '0');
  const hours = String(now.getHours()).padStart(2, '0');
  const minutes = String(now.getMinutes()).padStart(2, '0');
  const seconds = String(now.getSeconds()).padStart(2, '0');

  return time ? `${year}-${month}-${day} ${hours}:${minutes}:${seconds}` : `${year}-${month}-${day}`;
}

function toUTC(date, time) {
  let now = new Date(date);
  now.setMinutes(now.getMinutes() + now.getTimezoneOffset());
  return formatDate(now, time);
}

const PlotComponent = ({ title, data, defaultSlot = 'latest' }) => {
  const [slot, setSlot] = useState(defaultSlot);
  const theme = useTheme();
  const { primary, secondary } = theme.palette.text;
  const line = theme.palette.divider;
  const [series, setSeries] = useState([]);

  // chart options
  const areaChartOptions = {
    chart: {
      height: 450,
      type: 'area',
      toolbar: {
        show: false
      }
    },
    dataLabels: {
      enabled: false
    },
    stroke: {
      curve: 'smooth',
      width: 2
    },
    grid: {
      strokeDashArray: 0
    }
  };
  const [options, setOptions] = useState(areaChartOptions);

  useEffect(() => {
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

    const dataSeries = [];
    data.forEach((serie) => {
      dataSeries.push({
        name: serie.Label,
        dataAxis: xAxis.map((date) => {
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
      });
    });

    setOptions((prevState) => ({
      ...prevState,
      colors: [theme.palette.primary.main, theme.palette.secondary.main],
      xaxis: {
        categories: 
        slot === 'hourly'
          ? xAxis.map((date) => date.split(' ')[1])
          : xAxis,
        labels: {
          style: {
            fontSize: slot === 'latest' ? '0' : '11px',
          }
        },
        axisBorder: {
          show: true,
          color: line
        },
        tickAmount: xAxis.length,
      },
      yaxis: data.map((thisdata, ida) => ({
        opposite: ida === 1, 
        labels: {
          style: {
            colors: [secondary]
          },
          
          formatter: (num) => {
            if (Math.abs(num) >= 1e9) {
              return (num / 1e9).toFixed(1) + 'G'; // Convert to Millions
            } else if (Math.abs(num) >= 1e6) {
                return (num / 1e6).toFixed(1) + 'M'; // Convert to Millions
            } else if (Math.abs(num) >= 1e3) {
                return (num / 1e3).toFixed(1) + 'K'; // Convert to Thousands
            } else {
                return num.toString();
            }
          }
        },
        title: {
          text: thisdata.Label,
        }
      })),
      grid: {
        borderColor: line
      },
      tooltip: {
        theme: theme.palette.mode,
      }
    }));

    setSeries(dataSeries.map((serie) => {
      return {
        name: serie.name,
        data: serie.dataAxis
      }
    }));
  }, [slot, data]);


  return <Grid item xs={12} md={7} lg={8}>
    <Grid container alignItems="center" justifyContent="space-between">
      <Grid item>
        <Typography variant="h5">{title}</Typography>
      </Grid>
      <Grid item>
        <Stack direction="row" alignItems="center" spacing={0}>
          <Button
            size="small"
            onClick={() => setSlot('latest')}
            color={slot === 'latest' ? 'primary' : 'secondary'}
            variant={slot === 'latest' ? 'outlined' : 'text'}
          >
            Latest
          </Button>
          <Button
            size="small"
            onClick={() => setSlot('hourly')}
            color={slot === 'hourly' ? 'primary' : 'secondary'}
            variant={slot === 'hourly' ? 'outlined' : 'text'}
          >
            Hourly
          </Button>
          <Button
            size="small"
            onClick={() => setSlot('daily')}
            color={slot === 'daily' ? 'primary' : 'secondary'}
            variant={slot === 'daily' ? 'outlined' : 'text'}
          >
            Daily
          </Button>
        </Stack>
      </Grid>
    </Grid>
    <MainCard content={false} sx={{ mt: 1.5 }}>
      <Box sx={{ pt: 1, pr: 2 }}>
        <ReactApexChart options={options} series={series} type="area" height={450} />
      </Box>
    </MainCard>
  </Grid>
}

export default PlotComponent;