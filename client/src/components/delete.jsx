import { CheckOutlined, ClockCircleOutlined, DashboardOutlined, DeleteOutlined, DownOutlined, LockOutlined, SafetyOutlined, UpOutlined } from "@ant-design/icons";
import { Card, Chip, Stack, Tooltip, useTheme } from "@mui/material";
import { useState } from "react";

export const DeleteButton = ({onDelete, disabled, size}) => {
  const [confirmDelete, setConfirmDelete] = useState(false);

  return (<>
    {!confirmDelete && (<Chip label={<DeleteOutlined size={size}/>} 
      onClick={() => !disabled && setConfirmDelete(true)}/>)}
    {confirmDelete && (<Chip label={<CheckOutlined  size={size}/>} color="error" 
      onClick={(event) => !disabled && onDelete(event)}/>)}
  </>);
}