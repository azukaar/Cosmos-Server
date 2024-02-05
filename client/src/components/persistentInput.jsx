import { Checkbox } from "@mui/material";
import { useEffect } from "react";

export const PersistentCheckbox = ({ name, label, value, onChange }) => {
  let eKey = "PersistentCheckbox-" + name;

  useEffect(() => {
    if (localStorage.getItem(eKey) === null) {
      localStorage.setItem(eKey, value);
    } else {
      onChange(localStorage.getItem(eKey) === "true");
    }
  }, [name, value]);

  const handleChange = (e) => {
    onChange(e.target.checked);
    localStorage.setItem(eKey, e.target.checked);
  }

  return (
    <div style={{ display: 'flex', alignItems: 'center' }}>
      <input type="checkbox" name={name} hidden value={value} checked={value} />
      <Checkbox onChange={handleChange} checked={value} />
      {label}
    </div>
  );
}