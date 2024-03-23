import { Button, useMediaQuery, IconButton } from "@mui/material";


const ResponsiveButton = ({ children, startIcon, endIcon, size, style, ...props }) => {
  const isMobile = useMediaQuery((theme) => theme.breakpoints.down('sm'));
  let newStyle = style || {};
  if (isMobile) {
    newStyle.minHeight = '40px';
    newStyle.fontSize = '145%'
  }

  return (
    <Button 
      className="responsive-button"
      size={isMobile ? 'large' : size}
      startIcon={isMobile ? null : startIcon}
      endIcon={isMobile ? null : endIcon}
      {...props} style={newStyle}>
      {(isMobile) ? startIcon : (
        startIcon ? children : null
      )}
      {(isMobile) ? endIcon : (
        endIcon ? children : null
      )}
      {!startIcon && !endIcon ? children : null}
    </Button>
  );
}

export default ResponsiveButton;