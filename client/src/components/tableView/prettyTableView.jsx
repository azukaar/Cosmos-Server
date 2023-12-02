import * as React from 'react';
import {
  CircularProgress,
  Input,
  InputAdornment,
  Stack,
  TextField,
  useMediaQuery,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  useTheme,
} from '@mui/material';
import { SearchOutlined } from '@ant-design/icons';
import { Link } from 'react-router-dom';

const PrettyTableView = ({ isLoading, getKey, data, columns, sort, onRowClick, linkTo, buttons, fullWidth }) => {
  const [search, setSearch] = React.useState('');
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';
  const screenMin = {
    xs: useMediaQuery((theme) => theme.breakpoints.up('xs')),
    sm: useMediaQuery((theme) => theme.breakpoints.up('sm')),
    md: useMediaQuery((theme) => theme.breakpoints.up('md')),
    lg: useMediaQuery((theme) => theme.breakpoints.up('lg')),
    xl: useMediaQuery((theme) => theme.breakpoints.up('xl')),
    xxl: useMediaQuery((theme) => theme.breakpoints.up('xxl')),
  }

  return (
    <Stack direction="column" spacing={2} style={{width: fullWidth ? '100%': ''}}>
      <Stack direction="row" spacing={2}>
        <Input placeholder="Search"
          value={search}
          style={{
            width: '250px',
          }}
          startAdornment={
            <InputAdornment position="start">
              <SearchOutlined />
            </InputAdornment>
          }
          onChange={(e) => {
            setSearch(e.target.value);
          }}
        />
        {buttons}
      </Stack>
      
      {isLoading && (<div style={{height: '550px'}}>
                <center>
                    <br />
                    <CircularProgress />
                </center>
            </div>
            )}


      {!isLoading && <TableContainer style={{width: fullWidth ? '100%': '', background: isDark ? '#252b32' : '', borderTop: '3px solid ' + theme.palette.primary.main}} component={Paper}>
        <Table aria-label="simple table">
          <TableHead>
            <TableRow>
              {columns.map((column) => (
                (!column.screenMin || screenMin[column.screenMin]) && <TableCell>{column.title}</TableCell>
              ))}
            </TableRow>
          </TableHead>
          <TableBody>
            {data
              .filter((row) => {
                if (!search || search.length <= 2) return true;
                let found = false;
                columns.forEach((column) => {
                  if (column.search && column.search(row).toLowerCase().includes(search.toLowerCase())) {
                    found = true;
                  }
                })
                return found;
              })
              .sort((a, b) => {
                if (!sort) return 0;
                return sort(a, b);
              })
              .map((row, key) => (
                <TableRow
                  key={getKey(row)}
                  sx={{
                    cursor: 'pointer',
                    borderLeft: 'transparent solid 2px',
                    '&:last-child td, &:last-child th': { border: 0 },
                    '&:hover': {
                      backgroundColor: 'rgba(0, 0, 0, 0.06)',
                      borderColor: 'gray',
                      '&:hover .emphasis': {
                        textDecoration: 'underline'
                      }
                    },
                  }}
                >
                  {columns.map((column) => (
                
                    (!column.screenMin || screenMin[column.screenMin]) && <TableCell 
                      component={(linkTo && !column.clickable) ? Link : 'td'}
                      onClick={() => !column.clickable && onRowClick && onRowClick(row, key)}
                      to={linkTo && linkTo(row, key)}
                      className={column.underline ? 'emphasis' : ''}
                      sx={{
                        textDecoration: 'none',
                        ...column.style,
                      }}>
                        {column.field(row, key)}
                    </TableCell>
                  ))}
                </TableRow>
              ))}
          </TableBody>
        </Table>
      </TableContainer>}
    </Stack>
  )
}

export default PrettyTableView;