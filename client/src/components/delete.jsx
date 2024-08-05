import { CheckOutlined, ClockCircleOutlined, DashboardOutlined, DeleteOutlined, DownOutlined, LockOutlined, SafetyOutlined, UpOutlined } from "@ant-design/icons";
import { Card, Chip, IconButton, Stack, Tooltip } from "@mui/material";
import { useState } from "react";
import { useTheme } from '@mui/material/styles';

export const DeleteButton = ({onDelete, disabled, size}) => {
  const [confirmDelete, setConfirmDelete] = useState(false);

  return (<>
    {!confirmDelete && (<Chip label={<DeleteOutlined size={size} color={disabled ? 'silver' : 'error'}/>} 
      onClick={() => !disabled && setConfirmDelete(true)}/>)}
    {confirmDelete && (<Chip label={<CheckOutlined  size={size}/>} color={disabled ? 'silver' : 'error'}
      onClick={(event) => !disabled && onDelete(event)}/>)}
  </>);
}
export const DeleteIconButton = ({onDelete, disabled, size}) => {
  const [confirmDelete, setConfirmDelete] = useState(false);

  return (<>
    {!confirmDelete && (<IconButton color={disabled ? 'silver' : 'error'} onClick={() => !disabled && setConfirmDelete(true)}>
      <DeleteOutlined size={size} />
    </IconButton>)} 
    {confirmDelete && (<IconButton color="error" onClick={(event) => !disabled && onDelete(event)}>
      <CheckOutlined size={size} />
    </IconButton>)}
  </>);
}