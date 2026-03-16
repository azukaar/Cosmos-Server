import React from "react";
import { Tooltip } from "@mui/material";
import { useClientInfos } from "../utils/hooks";
import { useTranslation } from "react-i18next";

const PermissionGuard = ({ children, permission, alwaysShow }) => {
  const { t } = useTranslation();
  const { hasPermission, hasRolePermission } = useClientInfos();
  const hasPerm = hasPermission(permission);
  const hasRole = hasRolePermission(permission);

  if (!hasRole && !alwaysShow) return null;
  if (hasPerm) return children;

  return (
    <Tooltip title={t('sudo.required')}>
      <span style={{ cursor: 'not-allowed' }}>
        <span style={{ pointerEvents: 'none', filter: 'grayscale(0.5)', opacity: 0.65 }}>
          {React.cloneElement(children, { onClick: undefined, onChange: undefined })}
        </span>
      </span>
    </Tooltip>
  );
};

export default PermissionGuard;
