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

const ConfirmModal = ({ callback, label, content }) => {
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
              }}>Cancel</Button>
              <LoadingButton
              onClick={() => {   
                  callback();     
                  setOpenModal(false);    
              }}>Confirm</LoadingButton>
          </DialogActions>
      </Dialog>

      <Button
          disableElevation
          variant="outlined"
          color="warning"
          onClick={() => {
              setOpenModal(true);
          }}
      >
        {label}
      </Button>
    </>
};

export default ConfirmModal;
