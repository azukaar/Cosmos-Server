// material-ui
import * as React from 'react';
import { Alert, Button, Stack, Typography } from '@mui/material';
import { WarningOutlined, PlusCircleOutlined, CopyOutlined, ExclamationCircleOutlined , SyncOutlined, UserOutlined, KeyOutlined, ArrowUpOutlined, FileZipOutlined } from '@ant-design/icons';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Paper from '@mui/material/Paper';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogContentText from '@mui/material/DialogContentText';
import DialogTitle from '@mui/material/DialogTitle';
import TextField from '@mui/material/TextField';
import CircularProgress from '@mui/material/CircularProgress';
import Chip from '@mui/material/Chip';
import IconButton from '@mui/material/IconButton';
import * as API from '../../../api';
import MainCard from '../../../components/MainCard';
import IsLoggedIn from '../../../isLoggedIn';
import { useEffect, useState } from 'react';
import ResponsiveButton from '../../../components/responseiveButton';
import UploadButtons from '../../../components/fileUpload';
import NewDockerService from './newService';
import yaml from 'js-yaml';

function checkIsOnline() {
    API.isOnline().then((res) => {
        window.location.reload();
    }).catch((err) => {
        setTimeout(() => {
            checkIsOnline();
        }, 1000);
    });
}

const preStyle = {
  backgroundColor: '#000',
  color: '#fff',
  padding: '10px',
  borderRadius: '5px',
  overflow: 'auto',
  maxHeight: '500px',
  maxWidth: '100%',
  width: '100%',
  margin: '0',
  position: 'relative',
  fontSize: '12px',
  fontFamily: 'monospace',
  whiteSpace: 'pre-wrap',
  wordWrap: 'break-word',
  wordBreak: 'break-all',
  lineHeight: '1.5',
  boxShadow: '0 0 10px rgba(0,0,0,0.5)',
  border: '1px solid rgba(255,255,255,0.1)',
  boxSizing: 'border-box',
  marginBottom: '10px',
  marginTop: '10px',
  marginLeft: '0',
  marginRight: '0',
  display: 'block',
  textAlign: 'left',
  verticalAlign: 'baseline',
  opacity: '1',
}

const DockerComposeImport = () => {
    const [step, setStep] = useState(0);
    const [isLoading, setIsLoading] = useState(false);
    const [openModal, setOpenModal] = useState(false);
    const [dockerCompose, setDockerCompose] = useState('');
    const [service, setService] = useState({});

    useEffect(() => {
      if(dockerCompose === '') {
        return;
      }

      let newService = {};
      let doc = yaml.load(dockerCompose);

      // convert to the proper format
      if(doc.services) {
        Object.keys(doc.services).forEach((key) => {
          // convert volumes
          if(doc.services[key].volumes) {
            let volumes = [];
            doc.services[key].volumes.forEach((volume) => {
              let volumeSplit = volume.split(':');
              let volumeObj = {
                Source: volumeSplit[0],
                Target: volumeSplit[1],
                Type: volume[0] === '/' ? 'bind' : 'volume',
              };
              volumes.push(volumeObj);
            });
            doc.services[key].volumes = volumes;
          }

          // convert expose
          if(doc.services[key].expose) {
            doc.services[key].expose = doc.services[key].expose.map((port) => {
              return ''+port;
            })
          }
        });
      }

      setService(doc);
      console.log(doc);
    }, [dockerCompose]);

    return <>
        <Dialog open={openModal} onClose={() => setOpenModal(false)}>
            <DialogTitle>Import Docker Compose</DialogTitle>
            <DialogContent>
                <DialogContentText>
                  {step === 0 && <Stack spacing={2}>
                    <Alert severity="warning" icon={<WarningOutlined />}>
                      This is a highly experimental feature. It is recommended to use with caution.
                    </Alert>

                    <UploadButtons
                      accept='.yml,.yaml'
                      OnChange={(e) => {
                        const file = e.target.files[0];
                        const reader = new FileReader();
                        reader.onload = (e) => {
                          setDockerCompose(e.target.result);
                        };
                        reader.readAsText(file);
                      }}
                    />
                    
                    <TextField
                      multiline
                      placeholder='Paste your docker-compose.yml here or use the file upload button.'
                      fullWidth
                      value={dockerCompose}
                      onChange={(e) => setDockerCompose(e.target.value)}
                      style={preStyle}
                      rows={20}></TextField>
                  </Stack>}
                  {step === 1 && <Stack spacing={2}>
                    <NewDockerService service={service} />
                  </Stack>}
                </DialogContentText>
            </DialogContent>
            {!isLoading && <DialogActions>
                <Button onClick={() => {
                  setOpenModal(false);
                  setStep(0);
                  setDockerCompose('');
                }}>Close</Button>
                <Button onClick={() => {
                  if(step === 0) {
                    setStep(1);
                  } else {
                    setStep(0);
                  }
                }}>
                  {step === 0 && 'Next'}
                  {step === 1 && 'Back'}
                </Button>
            </DialogActions>}
        </Dialog>

        <ResponsiveButton
            color="primary"
            onClick={() => setOpenModal(true)}
            variant="outlined"
            startIcon={<ArrowUpOutlined />}
        >
            Import Docker Compose
        </ResponsiveButton>
    </>;
};

export default DockerComposeImport;
