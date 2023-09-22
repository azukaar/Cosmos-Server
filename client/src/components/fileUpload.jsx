import React from 'react';
import { Button } from '@mui/material';
import { UploadOutlined } from '@ant-design/icons';

export default function UploadButtons({OnChange, accept, label, variant, fullWidth, size}) {
  return (
    <div>
      <input
        accept={accept}
        style={{ display: 'none' }}
        id="contained-button-file"
        multiple
        type="file"
        onChange={OnChange}
      />
      <label htmlFor="contained-button-file">
        <Button variant={variant || "contained"} component="span"
        fullWidth={fullWidth} startIcon={<UploadOutlined />}>
          {label || 'Upload'}
        </Button>
      </label>
    </div>
  );
}