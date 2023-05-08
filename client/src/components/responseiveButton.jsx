import { Button, useMediaQuery, IconButton } from "@mui/material";


const ResponsiveButton = ({ children, startIcon, size, style, ...props }) => {
  const isMobile = useMediaQuery((theme) => theme.breakpoints.down('sm'));
  let newStyle = style || {};
  if (isMobile) {
    newStyle.minHeight = '40px';
    newStyle.fontSize = '145%'
  }

  return (
    <Button className="responsive-button" size={isMobile ? 'large' : size} startIcon={isMobile ? null : startIcon} {...props} style={newStyle}>
      {isMobile ? startIcon : children}
    </Button>
  );
}

export default ResponsiveButton;