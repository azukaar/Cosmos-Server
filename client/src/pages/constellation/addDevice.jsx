// material-ui
import { Alert, Button, Stack, TextField } from '@mui/material';
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
import { CosmosFormDivider, CosmosInputText, CosmosSelect } from '../config/users/formShortcuts';
import { DownloadFile } from '../../api/downloadButton';
import QRCode from 'qrcode';

const AddDeviceModal = ({ users, config, isAdmin, refreshConfig, devices }) => {
  const [openModal, setOpenModal] = useState(false);
  const [isDone, setIsDone] = useState(null);
  const canvasRef = React.useRef(null);

  let firstIP = "192.168.201.1/24";
  if (devices && devices.length > 0) { 
    const isIpFree = (ip) => {
      return devices.filter((d) => d.ip === ip).length === 0;
    }
    let i = 1;
    let j = 201;
    while (!isIpFree(firstIP)) {
      i++;
      if (i > 254) {
        i = 0;
        j++;
      }
      firstIP = "192.168." + j + "." + i + "/24";
    }
  }

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

  return <>
    <Dialog open={openModal} onClose={() => setOpenModal(false)}>
      <Formik
        initialValues={{
          nickname: users[0].nickname,
          deviceName: '',
          ip: firstIP,
          publicKey: '',
        }}

        validationSchema={yup.object({
        })}

        onSubmit={(values, { setSubmitting, setStatus, setErrors }) => {
          return API.constellation.addDevice(values).then(({data}) => {
            setIsDone(data);
            refreshConfig();
            renderCanvas(data);
          }).catch((err) => {
            setErrors(err.response.data);
          });
        }}
      >
        {(formik) => (
          <form onSubmit={formik.handleSubmit}>
            <DialogTitle>Add Device</DialogTitle>

            {isDone ? <DialogContent>
              <DialogContentText>
                <p>
                  Device added successfully!
                  Download scan the QR Code from the Cosmos app or download the relevant
                  files to your device along side the config and network certificate to
                  connect:
                </p>

                <Stack spacing={2} direction={"column"}>
                <CosmosFormDivider title={"Cosmos Client (QR Code)"} />
                <div style={{textAlign: 'center'}}>
                <canvas style={{borderRadius: '15px'}} ref={canvasRef} />
                </div>
                <CosmosFormDivider title={"Cosmos Client (File)"} />
                  <DownloadFile 
                    filename={isDone.DeviceName + `.constellation`}
                    content={JSON.stringify(isDone, null, 2)}
                    label={"Download " + isDone.DeviceName + `.constellation`}
                  />
                <CosmosFormDivider title={"Nebula Client"} />

                  <DownloadFile 
                    filename={`config.yml`}
                    content={isDone.Config}
                    label={"Download config.yml"}
                  />
                  <DownloadFile
                    filename={isDone.DeviceName + `.key`}
                    content={isDone.PrivateKey}
                    label={"Download " + isDone.DeviceName + `.key`}
                  />
                  <DownloadFile
                    filename={isDone.DeviceName + `.crt`}
                    content={isDone.PublicKey}
                    label={"Download " + isDone.DeviceName + `.crt`}
                  />
                  <DownloadFile
                    filename={`ca.crt`}
                    content={isDone.CA}
                    label={"Download ca.crt"}
                  />
                </Stack>
              </DialogContentText>
            </DialogContent> : <DialogContent>
              <DialogContentText>
                <p>Add a device to the constellation using either the Cosmos or Nebula client</p>
                <div>
                  <Stack spacing={2} style={{}}>
                    <CosmosSelect
                      name="nickname"
                      label="Owner"
                      formik={formik}
                      // disabled={!isAdmin}
                      options={
                        users.map((u) => {
                          return [u.nickname, u.nickname]
                        })
                      }
                    />

                    <CosmosInputText
                      name="deviceName"
                      label="Device Name"
                      formik={formik}
                    />

                    <CosmosInputText
                      name="ip"
                      label="Constellation IP Address"
                      formik={formik}
                    />

                    <CosmosInputText
                      multiline
                      name="publicKey"
                      label="Public Key (Optional)"
                      formik={formik}
                    />
                    <div>
                      {formik.errors && formik.errors.length > 0 && <Stack spacing={2} direction={"column"}>
                        <Alert severity="error">{formik.errors.map((err) => {
                          return <div>{err}</div>
                        })}</Alert>
                      </Stack>}
                    </div>
                  </Stack>
                </div>
              </DialogContentText>
            </DialogContent>}

            <DialogActions>
              <Button onClick={() => setOpenModal(false)}>Close</Button>
              <Button color="primary" variant="contained" type="submit">Add</Button>
            </DialogActions>
          </form>

        )}
      </Formik>
    </Dialog>

    <ResponsiveButton
      color="primary"
      onClick={() => {
        setIsDone(null);
        setOpenModal(true);
      }}
      variant="contained"
      startIcon={<PlusCircleFilled />}
    >
      Add Device
    </ResponsiveButton>
  </>;
};

export default AddDeviceModal;
