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
import { FormaterForMetric, toUTC } from './utils';


const PlotComponent = ({ title, slot, data, SimpleDesign, withSelector, xAxis, zoom, setZoom }) => {
  const theme = useTheme();
  const { primary, secondary } = theme.palette.text;
  const line = theme.palette.divider;
  const [series, setSeries] = useState([]);
  const [selected, setSelected] = useState(withSelector ? data.findIndex((d) => d.Key === withSelector)
     : 0);
     
  const isDark = theme.palette.mode === 'dark';

  const backColor = isDark ? '0,0,0' : '255,255,255';
  const textColor = isDark ? 'white' : 'dark';

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

    let toProcess = data;
    if(withSelector) {
      toProcess = data.filter((d, id) => id === selected);
    }
    const dataSeries = [];
    toProcess.forEach((serie) => {
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
            fontSize: (slot === 'latest' || SimpleDesign) ? '0' : '11px',
          }
        },
        axisBorder: {
          show: !SimpleDesign,
          color: line
        },
        tickAmount: xAxis.length,
        min: zoom.xaxis && zoom.xaxis.min,
        max: zoom.xaxis && zoom.xaxis.max,
      },
      yaxis: toProcess.map((thisdata, ida) => ({
        opposite: ida === 1, 
        labels: {
          style: {
            colors: [secondary],
            fontSize: '11px',
          },
          
          formatter: FormaterForMetric(thisdata)
        },
        title: {
          text: SimpleDesign ? '' : thisdata.Label,
        },
        min: zoom.yaxis && zoom.yaxis.min,
        max: zoom.yaxis && zoom.yaxis.max,
      })),
      grid: {
        borderColor: line
      },
      tooltip: {
        theme: theme.palette.mode,
      },
      chart: {
        events: {
          zoomed: function(chartContext, { xaxis, yaxis }) {
            // Handle the zoom event here
            console.log('Zoomed:', xaxis, yaxis);
            setZoom({ xaxis, yaxis });
          },
          selection: function(chartContext, { xaxis, yaxis }) {
            // Handle the selection event here
            console.log('Selected:', xaxis, yaxis);
          },
        }
      }
    }));

    setSeries(dataSeries.map((serie) => {
      return {
        name: serie.name,
        data: serie.dataAxis
      }
    }));
  }, [slot, data, selected]);

  useEffect(() => {
    // if different
    if(zoom && zoom.xaxis && zoom.yaxis && (!options.xaxis || !options.yaxis || !options.yaxis[0] || (
        zoom.xaxis.min !== options.xaxis.min ||
        zoom.xaxis.max !== options.xaxis.max ||
        zoom.yaxis.min !== options.yaxis[0].min ||
        zoom.yaxis.max !== options.yaxis[0].max 
    ))){
      setOptions((prevState) => ({
        ...prevState,
        xaxis: {
          ...prevState.xaxis,
          min: zoom.xaxis.min,
          max: zoom.xaxis.max,
        },
        yaxis: prevState.yaxis.map((y, id) => ({
          ...y,
          min: zoom.yaxis.min,
          max: zoom.yaxis.max,
        }))
      }));
    }
  }, [zoom]);

  return <Grid item xs={12} md={7} lg={8} >
    <Grid container alignItems="center" >
      <Grid item>
        <Typography variant="h5">{title}</Typography>
      </Grid>
      <Grid item >
          {withSelector && 
          <div style={{ marginLeft: 15 }}>
          <TextField
              id="standard-select-currency"
              size="small"
              select
              value={selected}
              onChange={(e) => setSelected(e.target.value)}
              sx={{ '& .MuiInputBase-input': { py: 0.5, fontSize: '0.875rem',  } }}
          >
              {data.map((option, i) => (
                  <MenuItem key={i} value={i}>
                      {option.Label}
                  </MenuItem>
              ))}
          </TextField></div>}
      </Grid>
    </Grid>
    <MainCard content={false} sx={{ mt: 1.5 }} >
      <Box sx={{ pt: 1, pr: 2 }}>
        <ReactApexChart options={options} series={series} type="area" height={450} />
      </Box>
    </MainCard>
  </Grid>
}

export default PlotComponent;