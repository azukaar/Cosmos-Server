// material-ui
import { Alert, Typography } from '@mui/material';
import { useState } from 'react';

// project import
import MainCard from '../../components/MainCard';

// ==============================|| SAMPLE PAGE ||============================== //

const ServeApps = () => {
  const {serveApps, setServeApps} = useState([]);

  return <div>
    <Alert severity="info">Implementation currently in progress! If you want to voice your opinion on where Cosmos is going, please join us on Discord!</Alert>
  </div>
}

export default ServeApps;
