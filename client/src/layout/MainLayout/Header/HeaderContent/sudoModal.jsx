import React, { useState } from 'react';
import { useFormik } from 'formik';
import { 
  Chip,
  Modal,
  Box,
  Typography,
  TextField,
  Button,
  CircularProgress,
  Alert
} from '@mui/material';
import { useClientInfos } from '../../../../utils/hooks';
import * as API from '../../../../api';
import { CosmosInputPassword } from '../../../../pages/config/users/formShortcuts';
import { useTranslation } from 'react-i18next';

const SudoModal = () => {
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const {userRole, role} = useClientInfos();
  const canSudo = role !== "2" && userRole === "2";
  const {t} = useTranslation();
  
  const formik = useFormik({
    initialValues: {
      password: ''
    },
    onSubmit: async (values, { resetForm }) => {
      try {
        setLoading(true);
        await API.auth.sudo(values);
        
        window.location.reload();
        
        setOpen(false);
        resetForm();
      } catch (error) {
        formik.setErrors({ password: 'Invalid password' });
      } finally {
        setLoading(false);
      }
    }
  });

  const handleOpen = () => setOpen(true);
  const handleClose = () => {
    setOpen(false);
    formik.resetForm();
  };

  return (
    canSudo ?
    <>
      <Chip 
        label="Admin" 
        onClick={handleOpen}
        color="primary"
        style={{height: '36px'}} 
      />
      <Modal
        open={open}
        onClose={handleClose}
      >
        <Box sx={{
          position: 'absolute',
          top: '50%',
          left: '50%',
          transform: 'translate(-50%, -50%)',
          width: 400,
          bgcolor: 'background.paper',
          borderRadius: 2,
          boxShadow: 24,
          p: 4
        }}>
          <Typography variant="h6" component="h2" gutterBottom>
            {t("sudo.title")}
          </Typography>
          <Alert severity="info" sx={{ mt: 2 }}>
            {t("sudo.description")}
          </Alert>
          <form onSubmit={formik.handleSubmit}>
            <CosmosInputPassword
              name="password"
              label="Password"
              noStrength
              formik={formik}
              autoComplete="current-password"
            />
            <Box sx={{ mt: 2, display: 'flex', justifyContent: 'flex-end', gap: 1 }}>
              <Button onClick={handleClose}>
                {t("global.cancelAction")}
              </Button>
              <Button 
                type="submit" 
                variant="contained" 
                disabled={loading}
              >
                {loading ? <CircularProgress size={24} /> : 'Submit'}
              </Button>
            </Box>
          </form>
        </Box>
      </Modal>
    </> :
    <></>
  );
};

export default SudoModal;