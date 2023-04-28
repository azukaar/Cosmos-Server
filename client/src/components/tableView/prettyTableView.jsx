import * as React from 'react';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Paper from '@mui/material/Paper';
import { Input, InputAdornment, Stack, TextField } from '@mui/material';
import { SearchOutlined } from '@ant-design/icons';
import { useTheme } from '@mui/material/styles';
import { Link } from 'react-router-dom';

const PrettyTableView = ({ getKey, data, columns, onRowClick, linkTo }) => {
  const [search, setSearch] = React.useState('');
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';

  return (
    <Stack direction="column" spacing={2}>
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

      <TableContainer style={{background: isDark ? '#252b32' : '', borderTop: '3px solid ' + theme.palette.primary.main}} component={Paper}>
        <Table sx={{ minWidth: 650 }} aria-label="simple table">
          <TableHead>
            <TableRow>
              {columns.map((column) => (
                <TableCell>{column.title}</TableCell>
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
              .map((row, key) => (
                <TableRow
                  onClick={() => onRowClick && onRowClick(row, key)}
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
                
                    <TableCell 
                    component={(linkTo && !column.clickable) ? Link : 'td'}
                    to={linkTo(row, key)}
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
      </TableContainer>
    </Stack>
  )
}

export default PrettyTableView;