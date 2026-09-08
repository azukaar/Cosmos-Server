const colors = {
  success: '#4caf50',
  warning: '#ffb300',
  orange: '#ff9800',
  error: '#f44336',
  unknown: '#9e9e9e',
};

const StatusDot = ({ status, size = 10, style, hollow }) => {
  const color = colors[status] || colors.unknown;
  return <span style={{
    display: 'inline-block',
    width: size,
    height: size,
    minWidth: size,
    borderRadius: '50%',
    boxSizing: 'border-box',
    ...(hollow ? { border: '2px solid ' + color } : { backgroundColor: color }),
    ...style,
  }} />;
};

export default StatusDot;
