import { Alert } from "@mui/material";
import { useTranslation } from "react-i18next";

export default function VMWarning() {
  const { t } = useTranslation();

  return (<Alert severity="warning">
    {t("mgmt.storage.vmWarning")}
  </Alert>);
}