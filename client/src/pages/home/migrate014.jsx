import React, { useState } from 'react';
import { FormikProvider, useFormik } from 'formik';
import * as Yup from 'yup';
import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  TextField,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Grid,
  FormHelperText,
  CircularProgress,
} from '@mui/material';
import { ArrowRightOutlined, PlusCircleOutlined } from '@ant-design/icons';
import { LoadingButton } from '@mui/lab';
import * as API from '../../api';

const Migrate014 = ({ config }) => {
  const [isOpened, setIsOpened] = useState(false);
  const [isSubmitted, setIsSubmitted] = useState(0);

  const formik = useFormik({
    initialValues: {
      http: config.HTTPConfig.HTTPPort || 80,
      https: config.HTTPConfig.HTTPSPort || 443,
    },
    enableReinitialize: true,
    validationSchema: Yup.object({
      http: Yup.number().required('HTTP port is required').typeError('HTTP port must be a number'),
      https: Yup.number().required('HTTPS port is required').typeError('HTTPS port must be a number'),
    }),
    onSubmit: (values, { setErrors, setStatus, setSubmitting }) => {
      return API.docker.migrateHost(values)
      .then((res) => {
          setSubmitting(true);
          setIsSubmitted(1);
          setTimeout(() => {
            setIsSubmitted(2);
          }, 15000);
          setStatus({ success: true });
          setSubmitting(false);
        }).catch((err) => {
          setStatus({ success: false });
          setErrors({ submit: err.message });
          setSubmitting(false);
        });
    },
  });

  return (
    <>
      <Dialog open={isOpened} onClose={() => setIsOpened(false)}>
        <FormikProvider value={formik}>
          <DialogTitle>Migrate to Host Mode</DialogTitle>
          {isSubmitted ? ((isSubmitted == 2) ?
          <DialogContent>
            <DialogContentText>
              Migration successful! Your Cosmos instance will now restart in Host Mode.
              Refresh the page to see the changes (it might take a few minutes!).
              if you are experiencing issues, please check the logs for more information.
            </DialogContentText>
          </DialogContent>
          :
          <DialogContent>
            <DialogContentText>
              <center>
                <CircularProgress
                  size={80}
                />
              </center>
            </DialogContentText>
          </DialogContent>)
          : 
          <DialogContent>
            <DialogContentText>
              In order to continue to improve your experience, Cosmos now supports the Host Mode of networking.
              It will be the recommended way to run your Cosmos container from now on. <strong>If you are using Windows or MacOS do not do this, it's not compatible</strong> (you can disable this warning in the config file).
              <br />Example of how it makes your life easier:
              <ul>
                <li>You will be able to connect to other services using localhost</li>
                <li>You won't need all the cosmos-network anymore to connect to containers</li>
                <li>You can see all your server's networking data in the monitoring tab (not just the docker data)</li>
                <li>Docker-compose imported container will work out-of-the-box</li>
                <li>Remove the docker overhead (small performance gain) and ensure Docker does not meddle with IP information</li>
                <li>Faster boot time (no bootstrapping of containers needed)</li>
              </ul>
              Cosmos can automatically migrate to the host mode, but before you do so, please confirm what ports you want to use with Cosmos (default are 80/443).
              The reason why we ask you to do this is that the host mode will use your server's network directly, not the docker virtual network. This means your port redirection (ex. -p 80:8080) will not be there anymore, and you need to tell Cosmos what ports to actually use directly. This form will save the settings for you and start the migration.
              If you have very customized Cosmos networking settings (ex. macvlan), the auto migration will not work, and you will need to do it manually.
              <br />
              <br />
            </DialogContentText>
            <form onSubmit={formik.handleSubmit}>
                <TextField
                id="http"
                name="http"
                fullWidth
                label="HTTP port"
                value={formik.values.http}
                onChange={formik.handleChange}
                error={formik.touched.http && Boolean(formik.errors.http)}
                helperText={formik.touched.http && formik.errors.http}
                style={{ marginBottom: '16px' }}
              />
              <TextField
                id="https"
                name="https"
                fullWidth
                label="HTTPS port"
                value={formik.values.https}
                onChange={formik.handleChange}
                error={formik.touched.https && Boolean(formik.errors.https)}
                helperText={formik.touched.https && formik.errors.https}
                style={{ marginBottom: '16px' }}
              />
            </form>
            {formik.errors.submit && (
              <Grid item xs={12}>
                <FormHelperText error>{formik.errors.submit}</FormHelperText>
              </Grid>
            )}
          </DialogContent>}
          {!isSubmitted && <DialogActions>
            <Button onClick={() => setIsOpened(false)}>Cancel</Button>
            <LoadingButton
              onClick={formik.handleSubmit}
              loading={formik.isSubmitting}
            >
              Migrate
            </LoadingButton>
          </DialogActions>}
        </FormikProvider>
      </Dialog>
      <div style={{ display: 'inline-block'}}>
        <Button
          style={{ margin: '2px 15px' }}
          onClick={() => setIsOpened(true)}
          variant="contained"
          startIcon={<ArrowRightOutlined />}
        >
          Help me migrate to Host Mode
        </Button>
      </div>
    </>
  );
};

export default Migrate014;