// material-ui
import { LoadingButton } from '@mui/lab';
import {
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogContentText,
    DialogTitle,
} from '@mui/material';
import * as React from 'react';
import { useEffect, useState } from 'react';

const ConfirmModal = ({ callback, label, content, startIcon }) => {
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
          startIcon={startIcon}
          onClick={() => {
              setOpenModal(true);
          }}
      >
        {label}
      </Button>
    </>
};

export default ConfirmModal;
