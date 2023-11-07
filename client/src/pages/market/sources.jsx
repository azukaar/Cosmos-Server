import * as React from 'react';
import IsLoggedIn from '../../isLoggedIn';
import * as API from '../../api';
import MainCard from '../../components/MainCard';
import { Formik, Field, useFormik, FormikProvider } from 'formik';
import * as Yup from 'yup';
import {
  Alert,
  Button,
  Checkbox,
  FormControlLabel,
  Grid,
  InputLabel,
  OutlinedInput,
  Stack,
  FormHelperText,
  TextField,
  MenuItem,
  Skeleton,
  CircularProgress,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogContentText,
  DialogActions,
  Box,
} from '@mui/material';
import { ContainerOutlined, ExclamationCircleOutlined, InfoCircleOutlined, PlusCircleOutlined, SyncOutlined, WarningOutlined } from '@ant-design/icons';
import PrettyTableView from '../../components/tableView/prettyTableView';
import { DeleteButton } from '../../components/delete';
import { CosmosCheckbox, CosmosFormDivider, CosmosInputText, CosmosSelect } from '../config/users/formShortcuts';
import ResponsiveButton from '../../components/responseiveButton';

const AlertValidationSchema = Yup.object().shape({
  name: Yup.string().required('Name is required'),
  trackingMetric: Yup.string().required('Tracking metric is required'),
  conditionOperator: Yup.string().required('Condition operator is required'),
  conditionValue: Yup.number().required('Condition value is required'),
  period: Yup.string().required('Period is required'),
});

const EditSourcesModal = ({ onSave }) => {
  const [config, setConfig] = React.useState(null);
  const [open, setOpen] = React.useState(false);
  
  function getConfig() {
    API.config.get().then((res) => {
      setConfig(res.data);
    });
  }

  React.useEffect(() => {
    getConfig();
  }, []);

  const formik = useFormik({
    initialValues: {
      sources: (config && config.MarketConfig.Sources) ? config.MarketConfig.Sources : [],
    },
    enableReinitialize: true, // This will reinitialize the form when `config` changes
    // validationSchema: AlertValidationSchema,
    onSubmit: (values) => {
      values.sources = values.sources.filter((a) => !a.removed);

      // setIsLoading(true);
  
      let toSave = {
        ...config,
        MarketConfig: {
          ...config.MarketConfig,
          Sources: values.sources,
        }
      };
  
      setOpen(false);

      return API.config.set(toSave).then(() => {
        onSave();
      });
    },
    validate: (values) => {
      const errors = {};

      values.sources && values.sources.forEach((source, index) => {
        if (source.removed) {
          return;
        }

        if (source.Name === '') {
          errors[`sources.${index}.Name`] = 'Name is required';
        }
        if (source.Url === '') {
          errors[`sources.${index}.Url`] = 'URL is required';
        }

        if (source.Name === 'cosmos-cloud' || values.sources.filter((s) => s.Name === source.Name && !s.removed).length > 1) {
          errors[`sources.${index}.Name`] = 'Name must be unique';
        }
      });

      console.log(errors)
  
      return errors;
    }
  });

  return (<>
    <Dialog open={open} onClose={() => setOpen(false)} maxWidth="sm" fullWidth>
      <DialogTitle>Edit Sources</DialogTitle>
      {config && <FormikProvider value={formik}>
      <form onSubmit={formik.handleSubmit}>
        <DialogContent>
          <Stack spacing={2}>
            {formik.values.sources && formik.values.sources
            .map((action, index) => {
              return !action.removed && <>
                <Stack spacing={0} key={index}>
                  <Stack direction="row" spacing={2}>
                    <CosmosInputText
                      name={`sources.${index}.Name`}
                      label="Name"
                      formik={formik}
                    />
                    <div style={{ flexGrow: 1 }}>
                      <CosmosInputText
                        name={`sources.${index}.Url`}
                        label="URL"
                        formik={formik}
                      />
                    </div>
                    
                    <Box  style={{
                      height: '95px',
                      display: 'flex',
                      alignItems: 'center',
                    }}>
                      <DeleteButton
                        onDelete={() => {
                          formik.setFieldValue(`sources.${index}.removed`, true);
                        }}
                      />
                    </Box>
                  </Stack>
                  <div>
                    <FormHelperText error>{formik.errors[`sources.${index}.Name`]}</FormHelperText>
                  </div>
                  <div>
                    <FormHelperText error>{formik.errors[`sources.${index}.Url`]}</FormHelperText>
                  </div>
                </Stack>
              </>
            })}

            <Button
              variant="outlined"
              color="primary"
              startIcon={<PlusCircleOutlined />}
              onClick={() => {
                formik.setFieldValue('sources', [
                  ...formik.values.sources,
                  {
                    Name: '',
                    Url: '',
                  },
                ]);
              }}>
              Add Source
            </Button>
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpen(false)}>Cancel</Button>
          <Button variant='contained' type="submit" disabled={formik.isSubmitting || !formik.isValid}>Save</Button>
        </DialogActions>
      </form>
      </FormikProvider>}
    </Dialog>


    <ResponsiveButton
      variant="outlined"
      startIcon={<ContainerOutlined />}
      onClick={() => setOpen(true)}
    >Sources</ResponsiveButton>
    </>
  );
};

export default EditSourcesModal;