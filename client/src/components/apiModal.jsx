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

const preStyle = {
  backgroundColor: '#000',
  color: '#fff',
  padding: '10px',
  borderRadius: '5px',
  overflow: 'auto',
  maxHeight: '520px',
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

const ApiModal = ({ callback, label }) => {
    const { t } = useTranslation();
    const [openModal, setOpenModal] = useState(false);
    const [content, setContent] = useState("");
    const [loading, setLoading] = useState(true);

    const getContent = async () => {
      setLoading(true);
      let content = await callback();
      setContent(content.data);
      setLoading(false);
    };

    useEffect(() => {
      if (openModal)
        getContent();
    }, [openModal]);

    return <>
      <Dialog open={openModal} onClose={() => setOpenModal(false)} fullWidth maxWidth={'sm'}>
          <DialogTitle>{t('global.refreshPage')}</DialogTitle>
          <DialogContent>
              <DialogContentText>
                <pre style={preStyle}>
                  {content}
                </pre>
              </DialogContentText>
          </DialogContent>
          <DialogActions>
              <LoadingButton
                  loading={loading}
              onClick={() => {   
                  getContent();         
              }}>{t('global.refresh')}</LoadingButton>
              <Button onClick={() => {
                  setOpenModal(false);           
              }}>Close</Button>
          </DialogActions>
      </Dialog>

      <Button
          disableElevation
          variant="outlined"
          color="primary"
          onClick={() => {
              setOpenModal(true);
          }}
      >
        {label}
      </Button>
    </>
};

export default ApiModal;
