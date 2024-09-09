// material-ui
import { Alert, Button, InputLabel, OutlinedInput, Stack, TextField } from '@mui/material';
import { LoadingButton } from '@mui/lab';
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
import { useTranslation } from 'react-i18next';

const AddDeviceModal = ({ users, config, refreshConfig, devices }) => {
  const { t } = useTranslation();
  const [openModal, setOpenModal] = useState(false);
  const [isDone, setIsDone] = useState(null);
  const canvasRef = React.useRef(null);
  const {role, nickname} = useClientInfos();
  const isAdmin = role === "2";

  let firstIP = "192.168.201.2/24";
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

    console.log("QRDATA", JSON.stringify(data))

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
          nickname: nickname,
          deviceName: '',
          ip: firstIP,
          publicKey: '',
          Port: "4242",
          PublicHostname: '',
          IsRelay: true,
          isLighthouse: false,
        }}

        validationSchema={yup.object({
        })}

        onSubmit={(values, { setSubmitting, setStatus, setErrors }) => {
          if(values.isLighthouse) values.nickname = null;

          return API.constellation.addDevice(values).then(({data}) => {
            setIsDone(data);
            refreshConfig();
            renderCanvas(data.Config);
          }).catch((err) => {
            setErrors(err.response.data);
          });
        }}
      >
        {(formik) => (
          <form onSubmit={formik.handleSubmit}>
            <DialogTitle>{t('mgmt.constellation.setup.addDeviceTitle')}</DialogTitle>

            {isDone ? <DialogContent>
              <DialogContentText>
                <p>
                {t('mgmt.constellation.setup.addDeviceSuccess')}
                </p>

                <Stack spacing={2} direction={"column"}>
                {/* {isDone.isLighthouse ? <>
                  <CosmosFormDivider title={"Docker"} />
                  <TextField
                    fullWidth
                    multiline
                    value={getDocker(isDone, false)}
                    variant="outlined"
                    size="small"
                    disabled
                  />
                  <CosmosFormDivider title={"File (Docker-Compose)"} />
                  <DownloadFile
                    filename={`docker-compose.yml`}
                    content={getDocker(isDone, true)}
                    label={"Download docker-compose.yml"}
                  />
                </> : <> */}
                  <CosmosFormDivider title={"QR Code"} />
                  <div style={{textAlign: 'center'}}>
                  <canvas style={{borderRadius: '15px'}} ref={canvasRef} />
                  </div>
                {/* </>} */}
                
                <CosmosFormDivider title={"File"} />
                  <DownloadFile 
                    filename={`constellation.yml`}
                    content={isDone.Config}
                    label={"Download constellation.yml"}
                  />
                </Stack>
              </DialogContentText>
            </DialogContent> : <DialogContent>
              <DialogContentText>
                <p>{t('mgmt.constellation.setup.addDeviceText')}</p>
                <div>
                  <Stack spacing={2} style={{}}>
                  <CosmosCheckbox
                    name="isLighthouse"
                    label="Lighthouse"
                    formik={formik}
                  />
                  {!formik.values.isLighthouse &&
                    (isAdmin ? <CosmosSelect
                      name="nickname"
                      label={t('mgmt.constellation.setup.owner.label')}
                      formik={formik}
                      // disabled={!isAdmin}
                      options={
                        users.map((u) => {
                          return [u.nickname, u.nickname]
                        })
                      }
                    /> : <>
                      <InputLabel>{t('mgmt.constellation.setup.owner.label')}</InputLabel>
                      <OutlinedInput
                        fullWidth
                        multiline
                        value={nickname}
                        variant="outlined"
                        size="small"
                        disabled
                      />
                    </>)}

                    <CosmosInputText
                      name="deviceName"
                      label={t('mgmt.constellation.setup.deviceName.label')}
                      formik={formik}
                    />

                    <CosmosInputText
                      name="ip"
                      label={t('mgmt.constellation.setup.ip.label')}
                      formik={formik}
                    />

                    {/* <CosmosInputText
                      name="Port"
                      label="VPN Port (default: 4242)"
                      formik={formik}
                    /> */}

                    {formik.values.isLighthouse && <>
                      <CosmosFormDivider title={t('mgmt.constellation.setuplighthouseTitle')} />

                      <CosmosInputText
                        name="PublicHostname"
                        label={t('mgmt.constellation.setup.pubHostname.label')}
                        formik={formik}
                      />

                      <CosmosCheckbox
                        name="IsRelay"
                        label={t('mgmt.constellation.isRelay.label')}
                        formik={formik}
                      />
                    </>}
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
              {!isDone && <LoadingButton 
              loading={formik.isSubmitting}
              color="primary" variant="contained" type="submit">{t('global.addAction')}</LoadingButton>}
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
      variant={
        "contained"
      }
      startIcon={<PlusCircleFilled />}
    >
      {t('mgmt.constellation.setup.addDeviceTitle')}
    </ResponsiveButton>
  </>;
};

export default AddDeviceModal;
