import { Stack } from '@mui/material';
import React from 'react';

function decodeUnicode(str) {
  return str.replace(/\\u([0-9a-zA-Z]{3-5})/g, (match, p1) => {
    return String.fromCharCode(parseInt(p1, 16));
  });
}

const LogLine = ({ message, docker, isMobile }) => {
  let html = decodeUnicode(message)
    .replace('\u0001\u0000\u0000\u0000\u0000\u0000\u0000', '')
    .replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;')
    .replace(/(?:\r\n|\r|\n)/g, '<br>')
    .replace(/�/g, '')
    .replace(/\x1b\[([0-9]{1,2}(?:;[0-9]{1,2})*)?m/g, (match, p1) => {
      if (!p1) {
        return '</span>';
      }
      const codes = p1.split(';');
      const styles = [];
      for (const code of codes) {
        switch (code) {
          case '1':
            styles.push('font-weight:bold');
            break;
          case '3':
            styles.push('font-style:italic');
            break;
          case '4':
            styles.push('text-decoration:underline');
            break;
          case '30':
          case '31':
          case '32':
          case '33':
          case '34':
          case '35':
          case '36':
          case '37':
          case '90':
          case '91':
          case '92':
          case '93':
          case '94':
          case '95':
          case '96':
          case '97':
            styles.push(`color:${getColor(code)}`);
            break;
        }
      }
      return `<span style="${styles.join(';')}">`;
    });

  if(docker) {
    let parts = html.match(/\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+Z/)
    if(!parts) {
      return <div dangerouslySetInnerHTML={{ __html: html }} />;
    }
    let restString = html.replace(parts[0], '')
    
    return <Stack direction={isMobile ? 'column' : 'row'} spacing={1}>
    <div style={{color:'grey', fontStyle:'italic', whiteSpace: 'pre'}}>
      {parts[0].replace('T', ' ').split('.')[0]}
    </div>
    <div dangerouslySetInnerHTML={{ __html: restString }} />
  </Stack>;
  }
    
  
  return <div dangerouslySetInnerHTML={{ __html: html }} />;
};

const getColor = (code) => {
  switch (code) {
    case '30':
    case '90':
      return 'black';
    case '31':
    case '91':
      return 'red';
    case '32':
    case '92':
      return 'green';
    case '33':
    case '93':
      return 'yellow';
    case '34':
    case '94':
      return 'blue';
    case '35':
    case '95':
      return 'magenta';
    case '36':
    case '96':
      return 'cyan';
    case '37':
    case '97':
      return 'white';
    default:
      return 'inherit';
  }
};

export default LogLine;