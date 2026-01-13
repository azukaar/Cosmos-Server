import React, { useId } from 'react';
import { Button } from '@mui/material';
import { UploadOutlined } from '@ant-design/icons';

export default function UploadButtons({OnChange, accept, label, variant, fullWidth, size}) {
  const id = useId();
  return (
    <div>
      <input
        accept={accept}
        style={{ display: 'none' }}
        id={id}
        multiple
        type="file"
        onChange={OnChange}
      />
      <label htmlFor={id}>
        <Button variant={variant || "contained"} component="span"
        fullWidth={fullWidth} startIcon={<UploadOutlined />}>
          {label || 'Upload'}
        </Button>
      </label>
    </div>
  );
}