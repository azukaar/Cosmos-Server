// material-ui
import { Alert, Button, InputLabel, OutlinedInput, Stack, TextField } from '@mui/material';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogContentText from '@mui/material/DialogContentText';
import DialogTitle from '@mui/material/DialogTitle';
import * as React from 'react';
import { useState } from 'react';
import ResponsiveButton from '../../components/responseiveButton';
import { PlusCircleFilled } from '@ant-design/icons';
import { Formik } from 'formik';
import * as yup from 'yup';
import * as API from '../../api';
import { CosmosCheckbox, CosmosFormDivider, CosmosInputText, CosmosSelect } from '../config/users/formShortcuts';
import { DownloadFile } from '../../api/downloadButton';
import QRCode from 'qrcode';
import { useClientInfos } from '../../utils/hooks';

const ResyncDeviceModal = ({ nickname, deviceName, OnClose }) => {
  const canvasRef = React.useRef(null);
  const [data, setData] = useState(null);

  const renderCanvas = (data) => {
    if (!canvasRef.current) return setTimeout(() => {
      renderCanvas(data);
    }, 500);

    QRCode.toCanvas(canvasRef.current, JSON.stringify(data), 
      {
        width: 600,
        color: {
          dark: "#000",
          light: '#fff'
        }
      }, function (error) {
        if (error) console.error(error)
      })
  }

  const getData = () => {
    return API.constellation.resyncDevice({nickname, deviceName}).then(({data}) => {
      renderCanvas(data);
      setData(data);
    })
  }

  React.useEffect(() => {
    getData();
  }, [nickname, deviceName]);

  return <>
    <Dialog open={true} onClose={OnClose}>
      <DialogTitle>Resync Device</DialogTitle>

      <DialogContent>
        <DialogContentText>
          <p>
            Use this to resync a client that lost connection to the server.
            In your client, click on "Resync Device" and follow the process.
            <strong>Do not use this on a new device, use the "Add Device" button instead.</strong>
          </p>

          <Stack spacing={2} direction={"column"}>
            <CosmosFormDivider title={"QR Code"} />
            <div style={{textAlign: 'center'}}>
              <canvas style={{borderRadius: '15px'}} ref={canvasRef} />
            </div>
          
            <CosmosFormDivider title={"File"} />
            <DownloadFile 
              filename={`constellation.resync.yml`}
              content={data}
              label={"Download constellation.resync.yml"}
            />
          </Stack>
        </DialogContentText>
      </DialogContent> 

      <DialogActions>
        <Button onClick={OnClose}>Done</Button>
      </DialogActions>
    </Dialog>
  </>;
};

export default ResyncDeviceModal;
