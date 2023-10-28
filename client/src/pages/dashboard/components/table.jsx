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
  TableHead,
  TableRow,
  TableCell,
  TableContainer,
  Table,
  TableBody
} from '@mui/material';
import MainCard from '../../../components/MainCard';

// material-ui
import { useTheme } from '@mui/material/styles';

// third-party
import ReactApexChart from 'react-apexcharts';
import { object } from 'prop-types';
import { FormaterForMetric } from './utils';

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

function descendingComparator(a, b, orderBy) {
  if (b[orderBy] < a[orderBy]) {
      return -1;
  }
  if (b[orderBy] > a[orderBy]) {
      return 1;
  }
  return 0;
}

function getComparator(order, orderBy) {
  return order === 'desc' ? (a, b) => descendingComparator(a, b, orderBy) : (a, b) => -descendingComparator(a, b, orderBy);
}

function stableSort(array, comparator) {
  const stabilizedThis = array.map((el, index) => [el, index]);
  stabilizedThis.sort((a, b) => {
      const order = comparator(a[0], b[0]);
      if (order !== 0) {
          return order;
      }
      return a[1] - b[1];
  });
  return stabilizedThis.map((el) => el[0]);
}

function toUTC(date, time) {
  let now = new Date(date);
  now.setMinutes(now.getMinutes() + now.getTimezoneOffset());
  return formatDate(now, time);
}

function OrderTableHead({ order, orderBy, headCells }) {
  return (
      <TableHead>
          <TableRow>
              {headCells.map((headCell) => (
                  <TableCell
                      key={headCell.id}
                      align={headCell.align}
                      padding={headCell.disablePadding ? 'none' : 'normal'}
                      sortDirection={orderBy === headCell.id ? order : false}
                  >
                      {headCell.label}
                  </TableCell>
              ))}
          </TableRow>
      </TableHead>
  );
}

const TableComponent = ({ title, data, defaultSlot = 'latest' }) => {
  const [slot, setSlot] = useState(defaultSlot);
  const theme = useTheme();
  const { primary, secondary } = theme.palette.text;
  const line = theme.palette.divider;
  const [series, setSeries] = useState([]);
  const [order] = useState('asc');
  const [orderBy] = useState('name');
  const [headCells, setHeadCells] = useState([]);
  const [rows, setRows] = useState([]);

  useEffect(() => {
    let heads = {};
    let fnrows = [];

    data.forEach((item) => {
      let k = item.Key.split('.')
      let v = item.Values.length ? item.Values[item.Values.length - 1] : 0;
      let name = k[k.length - 1];
      let cat = k[k.length - 2];

      heads[cat] = true;

      let formatter = FormaterForMetric(item); 

      if(!fnrows.find((row) => row.name === name)) {
        fnrows.push({
          name,
          [cat]: formatter(v.Value)
        });
      } else {
        fnrows.find((row) => row.name === name)[cat] = formatter(v.Value);
      }
    });

    // remove from fnrows rows where all values are 0
    fnrows = fnrows.filter((row) => {
      let flag = false;
      Object.keys(row).forEach((key) => {
        if(key !== 'name' && row[key] != 0) {
          flag = true;
        }
      });
      return flag;
    }
    );

    let fnhc = [
      {
          id: 'name',
          align: 'left',
          disablePadding: false,
          label: 'Name'
      },
    ];

    Object.keys(heads).forEach((head) => {
      fnhc.push({
        id: head,
        align: 'right',
        disablePadding: false,
        label: head
      });
    });

    setHeadCells(fnhc);
    setRows(fnrows);
    
  }, [data, slot]);


  return <Grid item xs={12} md={5} lg={4}>
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
        <Box>
            <TableContainer
                sx={{
                    width: '100%',
                    overflowX: 'auto',
                    position: 'relative',
                    display: 'block',
                    maxWidth: '100%',
                    '& td, & th': { whiteSpace: 'nowrap' }
                }}
            >
                <Table
                    aria-labelledby="tableTitle"
                    sx={{
                        '& .MuiTableCell-root:first-child': {
                            pl: 2
                        },
                        '& .MuiTableCell-root:last-child': {
                            pr: 3
                        }
                    }}
                >
                    <OrderTableHead headCells={headCells} order={order} orderBy={orderBy} />
                    <TableBody>
                        {stableSort(rows, getComparator(order, orderBy)).map((row, index) => {
                            const isItemSelected = false // isSelected(row.trackingNo);
                            const labelId = `enhanced-table-checkbox-${index}`;

                            return (
                                <TableRow
                                    hover
                                    role="checkbox"
                                    sx={{ '&:last-child td, &:last-child th': { border: 0 } }}
                                    aria-checked={isItemSelected}
                                    tabIndex={-1}
                                    key={row.trackingNo}
                                    selected={isItemSelected}
                                >
                                  {headCells.map((headCell) => {
                                    return <TableCell align={
                                      headCell.align
                                    }>{row[headCell.id]}</TableCell>
                                  })}
                                </TableRow>
                            );
                        })}
                    </TableBody>
                </Table>
            </TableContainer>
        </Box>
    </MainCard>
  </Grid>
}

export default TableComponent;