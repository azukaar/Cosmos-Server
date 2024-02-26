import { lazy, useEffect, useState } from 'react';
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
import Chart from 'react-apexcharts';
import { FormaterForMetric, toUTC } from './utils';


const PlotComponent = ({ title, slot, data, SimpleDesign, withSelector, xAxis, zoom, setZoom, zoomDisabled }) => {
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
      serie && dataSeries.push({
        name: serie.Label,
        dataAxis: xAxis.map((date) => {
          if(slot === 'latest') {
            // let old = serie.Values.length - 1 - date;
            // let realIndex = parseInt(serie.Values.length / xAxis.length  * date);
            // let currentIndex = serie.Values.length - 1 - realIndex;
            let index = serie.Values.length - 1 - parseInt(date / serie.TimeScale)
            
            return serie.Values[index] ? 
              serie.Values[index].Value :
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
      colors: [
        theme.palette.primary.main.replace('rgb(', 'rgba('), 
        theme.palette.secondary.main.replace('rgb(', 'rgba('),
        theme.palette.error.main.replace('rgb(', 'rgba('),
        theme.palette.warning.main.replace('rgb(', 'rgba('),
      ],
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
        opposite: ida % 2 === 1, 
        labels: {
          style: {
            colors: [secondary],
            fontSize: '11px',
          },
          
          formatter: FormaterForMetric(thisdata)
        },
        title: {
          text: thisdata && thisdata.Label,
        },
        max: (thisdata && thisdata.Max) ? thisdata.Max : undefined,
        min: (thisdata && thisdata.Max) ? 0 : undefined,
      })),
      grid: {
        borderColor: line
      },
      tooltip: {
        theme: theme.palette.mode,
      },
      chart: {
        zoom: {
          enabled: !zoomDisabled,
        },
        selection: {
          enabled: !zoomDisabled,
        },
        events: {
          zoomed: function(chartContext, { xaxis }) {
            setZoom({ xaxis });
          }
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
    if(zoom && zoom.xaxis && (!options.xaxis || (
        zoom.xaxis.min !== options.xaxis.min ||
        zoom.xaxis.max !== options.xaxis.max
    ))){
      setOptions((prevState) => ({
        ...prevState,
        xaxis: {
          ...prevState.xaxis,
          min: zoom.xaxis.min,
          max: zoom.xaxis.max,
        },
      }));
    }
  }, [zoom]);

  return <>
    <Grid container alignItems="center">
      <Grid item>
        <Typography variant="h5">{title}</Typography>
      </Grid>
      <Grid item>
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
      <Box sx={{ pt: 1, pr: 2}}>

        <Chart options={options} series={series} type="area" height={450} />
      </Box>
    </MainCard>
  </>
}

export default PlotComponent;