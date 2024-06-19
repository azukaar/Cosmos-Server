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
  TableBody,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogContentText,
  DialogActions
} from '@mui/material';
import MainCard from '../../../components/MainCard';

// material-ui
import { useTheme } from '@mui/material/styles';

// third-party
import { object } from 'prop-types';
import { FormaterForMetric } from './utils';
import { DownOutlined, UpOutlined } from '@ant-design/icons';
import PlotComponent from './plot';

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
  let a1 = a[orderBy];
  let b1 = b[orderBy];

  if(orderBy != 'name') {
    a1 = parseFloat(a["__" + orderBy]);
    b1 = parseFloat(b["__" + orderBy]);
  } else {
    a1 = a1.toLowerCase();
    b1 = b1.toLowerCase();
  }

  if (b1 < a1) {
      return -1;
  }
  if (b1 > a1) {
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

function OrderTableHead({ order, orderBy, headCells, setOrderBy, setOrder }) {
  return (
      <TableHead>
          <TableRow>
              {headCells.map((headCell) => (
                  <TableCell
                      key={headCell.id}
                      align={headCell.align}
                      padding={headCell.disablePadding ? 'none' : 'normal'}
                      sortDirection={orderBy === headCell.id ? order : false}
                      sx={{
                        '&:hover': {
                          cursor: 'pointer',
                          color: 'primary.main'
                        }
                      }}
                      onClick={() => {
                          console.log(headCell.id);
                          if(orderBy === headCell.id) {
                            setOrder(order === 'asc' ? 'desc' : 'asc');
                          } else {
                            setOrder('asc');
                            setOrderBy(headCell.id);
                          }
                      }}
                  >
                      {headCell.label} {orderBy === headCell.id ? (
                          <Box component="span" sx={{ ...(!headCell.disablePadding && { srOnly: true }) }}>
                              {order === 'desc' ? <DownOutlined /> : <UpOutlined />}
                          </Box>
                      ) : null}
                  </TableCell>
              ))}
          </TableRow>
      </TableHead>
  );
}

const TableComponent = ({ title, data, displayMax, render, xAxis, slot, zoom}) => {
  const theme = useTheme();
  const { primary, secondary } = theme.palette.text;
  const line = theme.palette.divider;
  const [series, setSeries] = useState([]);
  const [order, setOrder] = useState('asc');
  const [orderBy, setOrderBy] = useState('name');
  const [headCells, setHeadCells] = useState([]);
  const [rows, setRows] = useState([]);
  const [openModal, setOpenModal] = useState(false);

  useEffect(() => {
    let heads = {};
    let fnrows = [];

    data.forEach((item) => {
      let k = item.Key.split('.')
      let v = item.Values.length ? item.Values[item.Values.length - 1].Value : 0;

        let avgIndex = 0;
        v = xAxis
        .filter((date, index) => {
          if (zoom && zoom.xaxis && zoom.xaxis.min && zoom.xaxis.max) {
            return index >= zoom.xaxis.min && index <= zoom.xaxis.max;
          }
          return true;
        })
        .map((date) => {
          if (slot === 'hourly' || slot === 'daily') {
            let key = slot === 'hourly' ? "hour_" : "day_";
            let k = key + toUTC(date, slot === 'hourly');
            if (k in item.ValuesAggl) {
              return item.ValuesAggl[k].Value;
            } else {
              return 0;
            }
          } else {
            let realIndex = item.Values.length - 1 - parseInt(date / item.TimeScale)
            return item.Values[realIndex] ? item.Values[realIndex].Value : 0;
          }
        })
        .reduce((a, b) => {
          if (typeof b === "undefined" || b === null) {
            return a;
          }
          if (item.AggloType == "min") {
            return Math.min(a, b);
          } else if (item.AggloType == "max") {
            return Math.max(a, b);
          } else if (item.AggloType == "avg") {
            avgIndex++;
            return a + b;
          } else if (item.AggloType == "sum") {
            return a + b; 
          } else {
            return b;
          }
        }, 0);

        if (item.AggloType == "avg" && avgIndex) {
          v = v / avgIndex;
        }

      let name = k[k.length - 1];
      let cat = k[k.length - 2];

      heads[cat] = true;

      let formatter = FormaterForMetric(item, displayMax); 

      if(!fnrows.find((row) => row.name === name)) {
        fnrows.push({
          name,
          [cat]: render ? render(item, v, formatter(v)) : formatter(v),
          ["__" + cat]: v,
          "__key": [item.Key]
        });
      } else {
        fnrows.find((row) => row.name === name)[cat] = render ? render(item, v, formatter(v)) : formatter(v)
        fnrows.find((row) => row.name === name)["__" + cat] = v
        fnrows.find((row) => row.name === name)["__key"].push(item.Key)
      }
    });

    // remove from fnrows rows where all values are 0
    fnrows = fnrows.filter((row) => {
      let flag = false;
      Object.keys(row).forEach((key) => {
        if(key !== 'name' && row["__" + key]) {
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
    
  }, [data, slot, xAxis, zoom]);


  return <>
  {openModal && <Dialog open={openModal} onClose={() => setOpenModal(false)} maxWidth="md" fullWidth={true}>
        <DialogTitle>Detailed History</DialogTitle>
        <DialogContent>
          <DialogContentText  style={{
                }}>
              <PlotComponent 
                data={data.filter((item) => {
                  if (!openModal) {
                    return false;
                  }
                  return openModal.includes(item.Key)
                })}
                xAxis={xAxis}
                slot={slot}
                zoom={zoom}
                zoomDisabled={true}
              />
              </DialogContentText>
        </DialogContent>
        <DialogActions>
            <Button onClick={() => {
              setOpenModal(false)
            }}>Close</Button>
        </DialogActions>
    </Dialog>}
    
    <Grid item xs={12} md={5} lg={4}>
    <Grid container alignItems="center" justifyContent="space-between">
      <Grid item>
        <Typography variant="h5">{title}</Typography>
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
                    '& td, & th': { whiteSpace: 'nowrap' },
                    maxHeight: '474px'
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
                    <OrderTableHead headCells={headCells} order={order} orderBy={orderBy} setOrderBy={setOrderBy} setOrder={setOrder} />
                    <TableBody style={{height:'409px', overflow: 'auto'}}>
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
                                    onClick={() => {
                                      setOpenModal(row.__key);
                                    }}
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
  </>
}

export default TableComponent;