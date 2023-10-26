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

const PlotComponent = ({ data, defaultSlot = 'day' }) => {
  const [slot, setSlot] = useState(defaultSlot);
  const theme = useTheme();

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

  let hourlyDates = [];
  for(let i = 0; i < 48; i++) {
    let now = new Date();
    now.setHours(now.getHours() - i);
    now.setMinutes(0);
    now.setSeconds(0);
    // get as YYYY-MM-DD HH:MM:SS
    hourlyDates.unshift(formatDate(now, true));
  }

  let dailyDates = [];
  for(let i = 0; i < 30; i++) {
    let now = new Date();
    now.setDate(now.getDate() - i);
    dailyDates.unshift(formatDate(now));
  }

  const { primary, secondary } = theme.palette.text;
  const line = theme.palette.divider;

  const [options, setOptions] = useState(areaChartOptions);

  useEffect(() => {
    setOptions((prevState) => ({
      ...prevState,
      colors: [theme.palette.primary.main, theme.palette.secondary.main],
      xaxis: {
        categories:
          slot === 'hourly'
            ? hourlyDates.map((date) => date.split(' ')[1])
            : dailyDates,
        labels: {
          style: {
            fontSize: '11px',
          }
        },
        axisBorder: {
          show: true,
          color: line
        },
        tickAmount: slot === 'hourly' ? hourlyDates.length : dailyDates.length,
      },
      yaxis: [{
        labels: {
          style: {
            colors: [secondary]
          }
        },
        title: {
          text: data[0].Label,
        }
      },
      {
        opposite: true,
        labels: {
          style: {
            colors: [secondary]
          }
        },
        title: {
          text: data[1].Label,
        }
      }
      ],
      grid: {
        borderColor: line
      },
      tooltip: {
        theme: 'light'
      }
    }));
  }, [primary, secondary, line, theme, slot]);

  let dataSeries = [];
  data.forEach((serie) => {
    dataSeries.push({
      name: serie.Label,
      dataDaily: dailyDates.map((date) => {
        let k = "day_" + toUTC(date);
        if (k in serie.ValuesAggl) {
          return serie.ValuesAggl[k].Value;
        } else {
          console.log(k)
          return 0;
        }
      }),
      dataHourly: hourlyDates.map((date) => {
        let k = "hour_" + toUTC(date, true);
        if (k in serie.ValuesAggl) {
          return serie.ValuesAggl[k].Value;
        } else {
          return 0;
        }
      }),
    });
  });

  const [series, setSeries] = useState(dataSeries.map((serie) => {
    return {
      name: serie.name,
      data: slot === 'hourly' ? serie.dataHourly : serie.dataDaily
    }
  }));

  useEffect(() => {
    setSeries(dataSeries.map((serie) => {
      return {
        name: serie.name,
        data: slot === 'hourly' ? serie.dataHourly : serie.dataDaily
      }
    }));
  }, [slot, data]);


  return <Grid item xs={12} md={7} lg={8}>
    <Grid container alignItems="center" justifyContent="space-between">
      <Grid item>
        <Typography variant="h5">Server Resources</Typography>
      </Grid>
      <Grid item>
        <Stack direction="row" alignItems="center" spacing={0}>
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
      <Box sx={{ pt: 1, pr: 2 }} className="force-light">
        <ReactApexChart options={options} series={series} type="area" height={450} />;
      </Box>
    </MainCard>
  </Grid>
}

export default PlotComponent;