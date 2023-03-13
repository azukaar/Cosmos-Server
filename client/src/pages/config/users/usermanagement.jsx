// material-ui
import { Button, Typography } from '@mui/material';

import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Paper from '@mui/material/Paper';

import * as API from '../../../api';

// project import
import MainCard from '../../../components/MainCard';
import isLoggedIn from '../../../isLoggedIn';
import { useEffect, useState } from 'react';

// ==============================|| SAMPLE PAGE ||============================== //

const UserManagement = () => {

    const [rows, setRows] = useState([
        { id: 0, nickname: '12313', reset:'123132' },
        { id: 1, nickname: '354345', reset:'345345'  },
    ]);

    isLoggedIn();

    useEffect(() => {
        API.users.list()
            .then(data => {
                console.log(data);
                setRows(data.data);
            })
    }, [])

    return <MainCard title="Users">
        <TableContainer component={Paper}>
        <Table aria-label="simple table">
            <TableHead>
            <TableRow>
                <TableCell>Nickname</TableCell>
                <TableCell>Role</TableCell>
                <TableCell>Password</TableCell>
                <TableCell>Actions</TableCell>
            </TableRow>
            </TableHead>
            <TableBody>
            {rows.map((row) => (
                <TableRow
                    key={row.nickname}
                    sx={{ '&:last-child td, &:last-child th': { border: 0 } }}
                >
                <TableCell component="th" scope="row">
                    {row.nickname}
                </TableCell>
                <TableCell>User</TableCell>
                <TableCell><Button variant="contained" color="primary">Send Password Link</Button></TableCell>
                <TableCell><Button variant="contained" color="error">Delete</Button></TableCell>
                </TableRow>
            ))}
            </TableBody>
        </Table>
        </TableContainer>
    </MainCard>
};

export default UserManagement;
