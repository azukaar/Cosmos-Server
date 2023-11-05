import { CheckOutlined, ClockCircleOutlined, DashboardOutlined, DeleteOutlined, DownOutlined, LockOutlined, SafetyOutlined, UpOutlined } from "@ant-design/icons";
import { Card, Chip, Stack, Tooltip } from "@mui/material";
import { useState } from "react";
import { useTheme } from '@mui/material/styles';

export const DeleteButton = ({onDelete, disabled}) => {
  const [confirmDelete, setConfirmDelete] = useState(false);

  return (<>
    {!confirmDelete && (<Chip label={<DeleteOutlined />} 
      onClick={() => !disabled && setConfirmDelete(true)}/>)}
    {confirmDelete && (<Chip label={<CheckOutlined />} color="error" 
      onClick={(event) => !disabled && onDelete(event)}/>)}
  </>);
}