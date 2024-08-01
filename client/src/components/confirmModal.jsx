// material-ui
import { LoadingButton } from '@mui/lab';
import { Button } from '@mui/material';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogContentText from '@mui/material/DialogContentText';
import DialogTitle from '@mui/material/DialogTitle';
import * as React from 'react';
import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

const ConfirmModal = ({ callback, label, content, startIcon }) => {
    const { t } = useTranslation();
    const [openModal, setOpenModal] = useState(false);

    return <>
      <Dialog open={openModal} onClose={() => setOpenModal(false)}>
          <DialogTitle>Are you sure?</DialogTitle>
          <DialogContent>
              <DialogContentText>
                  {content}
              </DialogContentText>
          </DialogContent>
          <DialogActions>
              <Button onClick={() => {
                  setOpenModal(false);           
              }}>{t('global.cancelAction')}</Button>
              <LoadingButton
              onClick={() => {   
                  callback();     
                  setOpenModal(false);    
              }}>{t('global.confirmAction')}</LoadingButton>
          </DialogActions>
      </Dialog>

      <Button
          disableElevation
          variant="outlined"
          color="warning"
          startIcon={startIcon}
          onClick={() => {
              setOpenModal(true);
          }}
      >
        {label}
      </Button>
    </>
};


const ConfirmModalDirect = ({ callback, content, onClose }) => {
    const [openModal, setOpenModal] = useState(true);

    return <>
      <Dialog open={openModal} onClose={() => {
        onClose && onClose();
        setOpenModal(false);
      }}>
          <DialogTitle>{t('global.confirmDeletion')}</DialogTitle>
          <DialogContent>
              <DialogContentText>
                  {content}
              </DialogContentText>
          </DialogContent>
          <DialogActions>
              <Button onClick={() => {
                  setOpenModal(false);    
                  onClose && onClose();
              }}>{t('global.cancelAction')}</Button>
              <LoadingButton
              onClick={() => {   
                  callback();     
                  setOpenModal(false);    
                  onClose && onClose();
              }}>{t('global.confirmAction')}</LoadingButton>
          </DialogActions>
      </Dialog>
    </>
};

export default ConfirmModal;
export { ConfirmModalDirect };
