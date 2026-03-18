import React, { useRef, useState, useEffect, cloneElement } from 'react';
import { Tooltip } from '@mui/material';

function Ellipsis({ children, title, maxWidth, style, ...props }) {
  const ref = useRef(null);
  const [isOverflowing, setIsOverflowing] = useState(false);

  useEffect(() => {
    const el = ref.current;
    if (el) {
      // Check the element itself or its first child for overflow
      const target = el.querySelector('.MuiChip-label') || el;
      setIsOverflowing(target.scrollWidth > target.clientWidth);
    }
  });

  const ellipsisStyle = {
    maxWidth: maxWidth || '100%',
    ...style,
  };

  const child = cloneElement(children, {
    ref,
    style: { ...children.props.style, ...ellipsisStyle },
    ...props,
  });

  if (!isOverflowing) return child;

  return (
    <Tooltip title={title || children} arrow>
      {child}
    </Tooltip>
  );
}

export default Ellipsis;
