// material-ui
import * as React from 'react';
import { Alert, Button, Stack, Typography } from '@mui/material';
import { WarningOutlined, PlusCircleOutlined, CopyOutlined, ExclamationCircleOutlined , SyncOutlined, UserOutlined, KeyOutlined } from '@ant-design/icons';
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
import { useEffect, useState } from 'react';
import { isDomain } from '../../../utils/indexs';
import { useTranslation } from 'react-i18next';

function checkIsOnline() {
    API.isOnline().then((res) => {
        window.location.reload();
    }).catch((err) => {
        setTimeout(() => {
            checkIsOnline();
        }, 1000);
    });
}

const RestartModal = ({openModal, setOpenModal, config, newRoute, isHostMachine }) => {
    const { t } = useTranslation();
    const [isRestarting, setIsRestarting] = useState(false);
    const [warn, setWarn] = useState(false);
    const needsRefresh = config && (config.HTTPConfig.HTTPSCertificateMode == "SELFSIGNED" ||
        !isDomain(config.HTTPConfig.Hostname))
    const isNotDomain = config && !isDomain(config.HTTPConfig.Hostname);
    let newRouteWarning = config && (config.HTTPConfig.HTTPSCertificateMode == "LETSENCRYPT" && newRoute && 
        (!config.HTTPConfig.DNSChallengeProvider || !config.HTTPConfig.UseWildcardCertificate))
    const [errorRestart, setErrorRestart] = useState(null);

    return config ? (<>
        {needsRefresh && <>
            <Dialog open={openModal} onClose={() => setOpenModal(false)}>
                <DialogTitle>{t('global.refreshPage')}</DialogTitle>
                <DialogContent>
                    <DialogContentText>
                        {t('mgmt.config.proxy.refreshNeededWarning.selfSigned')} {isNotDomain && t('mgmt.config.proxy.refreshNeededWarning.notDomain')}
                    </DialogContentText>
                </DialogContent>
                <DialogActions>
                    <Button onClick={() => {
                        window.location.reload(true);                
                    }}>{t('global.refresh')}</Button>
                </DialogActions>
            </Dialog>
        </>}
        {newRouteWarning && <>
            <Dialog open={openModal} onClose={() => setOpenModal(false)}>
                <DialogTitle>{t('mgmt.config.certRenewalTitle')}</DialogTitle>
                <DialogContent>
                    <DialogContentText>
                        {t('mgmt.config.certRenewalText')} <a target="_blank" rel="noopener noreferrer" href="https://cosmos-cloud.io/doc/9%20Other%20Setups/#dns-challenge-and-wildcard-certificates">{t('mgmt.config.certRenewalLinktext')}</a>.
                    </DialogContentText>
                </DialogContent>
                <DialogActions>
                    <Button onClick={() => {
                        setOpenModal(false);             
                    }}>{t('mgmt.config.restart.okButton')}</Button>
                </DialogActions>
            </Dialog>
        </>}
    </>)
    :(<>
        <Dialog open={openModal} onClose={() => setOpenModal(false)}>
            <DialogTitle>{!isRestarting ? t('mgmt.config.restart.restartTitle') : t('mgmt.config.restart.restartStatus')}</DialogTitle>
            <DialogContent>
                <DialogContentText>
                    {errorRestart && <Alert severity="error" icon={<ExclamationCircleOutlined />}>
                        {errorRestart}
                    </Alert>}
                    {warn && !errorRestart  && <div>
                        <Alert severity="warning" icon={<WarningOutlined />}>
                        {t('mgmt.config.restart.restartTimeoutWarning')}<br />{t('mgmt.config.restart.restartTimeoutWarningTip')}
                        </Alert>
                    </div>}
                    {isRestarting && !errorRestart && 
                    <div style={{textAlign: 'center', padding: '20px'}}>
                        <CircularProgress />
                    </div>}
                    {!isRestarting && t('mgmt.config.restart.restartQuestion')}
                </DialogContentText>
            </DialogContent>
            {!isRestarting && <DialogActions>
                <Button onClick={() => setOpenModal(false)}>{t('mgmt.config.restart.laterButton')}</Button>
                <Button onClick={() => {
                    setIsRestarting(true);
                    setErrorRestart(null);

                    if(isHostMachine) {
                        API.restartServer()
                        .then((res) => {
                        }).catch((err) => {
                            setErrorRestart(err.message);
                        });
                    } else {
                        API.config.restart()
                    }
                    
                    setTimeout(() => {
                        if(!errorRestart) {
                            checkIsOnline();
                        }
                    }, 1500);

                    setTimeout(() => {
                        setWarn(true);
                    }, 20000)
                }}>{t('mgmt.servapps.actionBar.restart')}</Button>
            </DialogActions>}
        </Dialog>
    </>);
};

export default RestartModal;
