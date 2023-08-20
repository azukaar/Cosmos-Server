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
import { CosmosInputText, CosmosSelect } from '../config/users/formShortcuts';
import { DownloadFile } from '../../api/downloadButton';

const AddDeviceModal = ({ config, isAdmin, refreshConfig, devices }) => {
  const [openModal, setOpenModal] = useState(false);
  const [isDone, setIsDone] = useState(null);

  return <>
    <Dialog open={openModal} onClose={() => setOpenModal(false)}>
      <Formik
        initialValues={{
          nickname: '',
          deviceName: '',
          ip: '192.168.201.1/24',
          publicKey: '',
        }}

        validationSchema={yup.object({
        })}

        onSubmit={(values, { setSubmitting, setStatus, setErrors }) => {
          return API.constellation.addDevice(values).then(({data}) => {
            setIsDone(data);
            refreshConfig();
          }).catch((err) => {
            setErrors(err.response.data);
          });
        }}
      >
        {(formik) => (
          <form onSubmit={formik.handleSubmit}>
            <DialogTitle>Manually Add Device</DialogTitle>

            {isDone ? <DialogContent>
              <DialogContentText>
                <p>
                  Device added successfully!
                  Download the private and public keys to your device along side the config and network certificate to connect:
                </p>

                <Stack spacing={2} direction={"column"}>
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
                <p>Manually add a device to the constellation. It is recommended that you use the Cosmos app instead. Use this form to add another Nebula device manually</p>
                <div>
                  <Stack spacing={2} style={{}}>
                    <CosmosSelect
                      name="nickname"
                      label="Owner"
                      formik={formik}
                      // disabled={!isAdmin}
                      options={[
                        ["admin", "admin"]
                      ]}
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
      Manually Add Device
    </ResponsiveButton>
  </>;
};

export default AddDeviceModal;
