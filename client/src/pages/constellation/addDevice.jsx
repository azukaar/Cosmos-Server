// material-ui
import { Alert, Button, InputLabel, OutlinedInput, Stack, TextField, Tooltip } from '@mui/material';
import { LoadingButton } from '@mui/lab';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogContentText from '@mui/material/DialogContentText';
import DialogTitle from '@mui/material/DialogTitle';
import * as React from 'react';
import { useState } from 'react';
import ResponsiveButton from '../../components/responseiveButton';
import { PlusCircleFilled, QuestionCircleOutlined } from '@ant-design/icons';
import { Formik } from 'formik';
import * as yup from 'yup';
import * as API from '../../api';
import { CosmosCheckbox, CosmosFormDivider, CosmosInputText, CosmosSelect } from '../config/users/formShortcuts';
import { DownloadFile } from '../../api/downloadButton';
import QRCode from 'qrcode';
import { useClientInfos } from '../../utils/hooks';
import { useTranslation } from 'react-i18next';
import { json } from 'react-router';

const AddDeviceModal = ({ users, config, refreshConfig, devices, canCreateManager, canCreateAgent }) => {
  const { t } = useTranslation();
  const [openModal, setOpenModal] = useState(false);
  const [isDone, setIsDone] = useState(null);
  const [nextIP, setNextIP] = useState("");
  const canvasRef = React.useRef(null);
  const {role, nickname} = useClientInfos();
  const isAdmin = role === "2";

  const fetchNextIP = async () => {
    try {
      const res = await API.constellation.getNextIP();
      if (res.data) {
        setNextIP(res.data);
      }
    } catch (err) {
      console.error("Error fetching next IP:", err);
    }
  };

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
        enableReinitialize
        initialValues={{
          nickname: nickname,
          deviceName: '',
          ip: nextIP,
          publicKey: '',
          Port: "4242",
          PublicHostname: '',
          IsRelay: true,
          IsExitNode: true,
          IsLoadBalancer: true,
          deviceType: 'client',
          isLighthouse: true,
          invisible: false,
        }}

        validationSchema={yup.object({
          deviceName: yup.string().required().min(3).max(32)
            .matches(/^[a-z0-9-]+$/, t('mgmt.constellation.setup.deviceName.validationError')),
        })}

        onSubmit={(values, { setSubmitting, setStatus, setErrors }) => {
          const isCosmosServer = values.deviceType === 'cosmos-manager' || values.deviceType === 'cosmos-agent';
          const isLighthouse = values.deviceType === 'lighthouse' || (isCosmosServer && values.isLighthouse);
          const cosmosNode = values.deviceType === 'cosmos-manager' ? 2 : values.deviceType === 'cosmos-agent' ? 1 : 0;

          if(isLighthouse) values.nickname = null;

          const payload = {
            ...values,
            isLighthouse,
            cosmosNode,
          };

          return API.constellation.addDevice(payload).then(({data}) => {
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
                  <CosmosSelect
                    name="deviceType"
                    label={<Stack direction="row" spacing={0.5} alignItems="center" component="span">
                      <span>{t('mgmt.constellation.setup.deviceType.label')}</span>
                      <Tooltip title={t('mgmt.constellation.setup.deviceType.tooltip')}>
                        <QuestionCircleOutlined style={{ fontSize: 14, cursor: 'help', opacity: 0.6 }} />
                      </Tooltip>
                    </Stack>}
                    formik={formik}
                    disabled={!isAdmin}
                    options={isAdmin ? [
                      ['client', t('mgmt.constellation.setup.deviceType.client')],
                      ['lighthouse', t('mgmt.constellation.setup.deviceType.lighthouse')],
                      ['cosmos-agent', t('mgmt.constellation.setup.deviceType.cosmosAgent'), !canCreateAgent],
                      ['cosmos-manager', t('mgmt.constellation.setup.deviceType.cosmosManager'), !canCreateManager],
                    ] : [
                      ['client', t('mgmt.constellation.setup.deviceType.client')],
                    ]}
                  />
                  {formik.values.deviceType === 'client' &&
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

                    {formik.values.deviceType === 'client' && <>
                      <CosmosCheckbox
                        name="invisible"
                        label={t('mgmt.constellation.setup.invisible.label')}
                        formik={formik}
                      />
                    </>}

                    {(formik.values.deviceType === 'lighthouse' || formik.values.deviceType === 'cosmos-manager' || formik.values.deviceType === 'cosmos-agent') && <>
                      <CosmosFormDivider title={t('mgmt.constellation.setuplighthouseTitle')} />

                      {(formik.values.deviceType === 'cosmos-manager' || formik.values.deviceType === 'cosmos-agent') && <>
                        <CosmosCheckbox
                          name="isLighthouse"
                          label={<Stack direction="row" spacing={0.5} alignItems="center" component="span">
                            <span>{t('mgmt.constellation.setup.isLighthouse.label')}</span>
                            <Tooltip title={t('mgmt.constellation.setup.isLighthouse.tooltip')}>
                              <QuestionCircleOutlined style={{ fontSize: 14, cursor: 'help', opacity: 0.6 }} />
                            </Tooltip>
                          </Stack>}
                          formik={formik}
                        />
                      </>}

                      {formik.values.isLighthouse && <CosmosInputText
                        name="PublicHostname"
                        label={t('mgmt.constellation.setup.pubHostname.label')}
                        formik={formik}
                      />}

                      {(formik.values.deviceType === 'cosmos-manager' || formik.values.deviceType === 'cosmos-agent') && <>
                        <CosmosCheckbox
                          name="IsLoadBalancer"
                          label={<Stack direction="row" spacing={0.5} alignItems="center" component="span">
                            <span>{t('mgmt.constellation.isLoadBalancer.label')}</span>
                            <Tooltip title={t('mgmt.constellation.setup.loadBalancer.tooltip')}>
                              <QuestionCircleOutlined style={{ fontSize: 14, cursor: 'help', opacity: 0.6 }} />
                            </Tooltip>
                          </Stack>}
                          formik={formik}
                        />
                      </>}

                      <CosmosCheckbox
                        name="IsRelay"
                        label={<Stack direction="row" spacing={0.5} alignItems="center" component="span">
                          <span>{t('mgmt.constellation.isRelay.label')}</span>
                          <Tooltip title={t('mgmt.constellation.setup.relayRequests.tooltip')}>
                            <QuestionCircleOutlined style={{ fontSize: 14, cursor: 'help', opacity: 0.6 }} />
                          </Tooltip>
                        </Stack>}
                        formik={formik}
                      />

                      <CosmosCheckbox
                        name="IsExitNode"
                        label={<Stack direction="row" spacing={0.5} alignItems="center" component="span">
                          <span>{t('mgmt.constellation.isExitNode.label')}</span>
                          <Tooltip title={t('mgmt.constellation.setup.exitNode.tooltip')}>
                            <QuestionCircleOutlined style={{ fontSize: 14, cursor: 'help', opacity: 0.6 }} />
                          </Tooltip>
                        </Stack>}
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
        fetchNextIP();
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
