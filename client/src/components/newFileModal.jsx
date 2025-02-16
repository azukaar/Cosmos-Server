// material-ui
import * as React from 'react';
import { Alert, Box, Button, Checkbox, CircularProgress, FormControl, FormHelperText, Grid, Icon, IconButton, InputLabel, List, ListItem, ListItemButton, MenuItem, NativeSelect, Select, Stack, TextField, Typography } from '@mui/material';
import Dialog from '@mui/material/Dialog';
import {TreeItem, TreeView} from '@mui/x-tree-view';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogContentText from '@mui/material/DialogContentText';
import DialogTitle from '@mui/material/DialogTitle';
import { LoadingButton } from '@mui/lab';
import { useFormik, FormikProvider } from 'formik';
import * as Yup from 'yup';
import { useTranslation } from 'react-i18next';
import { FolderOpenOutlined, ClockCircleOutlined, DashboardOutlined, DeleteOutlined, DownOutlined, LockOutlined, SafetyOutlined, UpOutlined, ExpandOutlined, FolderFilled, FileFilled, FileOutlined, LoadingOutlined, FolderAddFilled, FolderViewOutlined } from "@ant-design/icons";
import * as API from '../api';
import { current } from '@reduxjs/toolkit';
import { json } from 'react-router';

const NewFolderModal = ({ raw, open, cb, OnClose, onPick, canCreate, storage = '', path = '/', select = 'any' }) => {
  const { t } = useTranslation();

  const formik = useFormik({
    initialValues: {
      name: '',
    },
    onSubmit: (values) => {
      API.storage.newFolder(storage, path, values.name).then((res) => {
        OnClose();
        cb && cb(res.data.created, values.name);
      });
    },
  });

  return (
    <Dialog open={open} onClose={OnClose} fullWidth>
      <form onSubmit={formik.handleSubmit}>
        <DialogTitle>{t('mgmt.storage.newFolderIn')} {storage}:{path}</DialogTitle>
        <DialogContent>
          <DialogContentText>
            <TextField
              autoFocus
              margin="dense"
              name='name'
              label={t('mgmt.storage.newFolderName')}
              type="text"
              fullWidth
              value={formik.values.name}
              onChange={formik.handleChange}
              onBlur={formik.handleBlur}
            />
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={OnClose}>{t('global.cancelAction')}</Button>
          <Button autoFocus type="submit">
            {t('global.confirmAction')}
          </Button>
        </DialogActions>
      </form>
    </Dialog>
  );
};

const NewFolderButton = ({ cb, storage = '', path = '/' }) => {
  const [open, setOpen] = React.useState(false);
  const { t } = useTranslation();

  return (
    <>
      <Button variant="contained" onClick={() => setOpen(true)}>{t('mgmt.storage.newFolder')}</Button>
      {open && <NewFolderModal cb={cb}  open={open} OnClose={() => setOpen(false)} storage={storage} path={path} />}
    </>
  );
}

export default NewFolderModal;
export { NewFolderButton };