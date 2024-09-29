import { Stack } from '@mui/material';
import React from 'react';

function decodeUnicode(str) {
  return str.replace(/\\u([0-9a-zA-Z]{3-5})/g, (match, p1) => {
    return String.fromCharCode(parseInt(p1, 16));
  });
}

const processANSISequences = (input) => {
  const lines = [''];
  let cursorX = 0;
  let cursorY = 0;

  const moveCursor = (x, y) => {
    cursorX = Math.max(0, x);
    cursorY = Math.max(0, y);
    while (cursorY >= lines.length) {
      lines.push('');
    }
  };

  const eraseLine = (mode) => {
    const line = lines[cursorY] || '';
    switch (mode) {
      case 0: // Erase from cursor to end of line
        lines[cursorY] = line.slice(0, cursorX);
        break;
      case 1: // Erase from start of line to cursor
        lines[cursorY] = ' '.repeat(cursorX) + line.slice(cursorX);
        break;
      case 2: // Erase entire line
        lines[cursorY] = '';
        break;
    }
  };

  for (let i = 0; i < input.length; i++) {
    if (input[i] === '\x1b' && input[i + 1] === '[') {
      i += 2;
      let sequence = '';
      while (i < input.length && !'ABCDEFGHJKSTfmnu'.includes(input[i])) {
        sequence += input[i++];
      }
      const command = input[i];

      if (command === 'A') { // Cursor Up
        moveCursor(cursorX, cursorY - parseInt(sequence || '1'));
      } else if (command === 'B') { // Cursor Down
        moveCursor(cursorX, cursorY + parseInt(sequence || '1'));
      } else if (command === 'C') { // Cursor Forward
        moveCursor(cursorX + parseInt(sequence || '1'), cursorY);
      } else if (command === 'D') { // Cursor Back
        moveCursor(cursorX - parseInt(sequence || '1'), cursorY);
      } else if (command === 'E') { // Cursor Next Line
        moveCursor(0, cursorY + parseInt(sequence || '1'));
      } else if (command === 'F') { // Cursor Previous Line
        moveCursor(0, cursorY - parseInt(sequence || '1'));
      } else if (command === 'G') { // Cursor Horizontal Absolute
        moveCursor(parseInt(sequence) - 1, cursorY);
      } else if (command === 'H') { // Cursor Position
        const [y, x] = sequence.split(';').map(n => parseInt(n) - 1);
        moveCursor(x || 0, y || 0);
      } else if (command === 'J') { // Erase in Display
        const mode = parseInt(sequence || '0');
        if (mode === 0) {
          lines[cursorY] = lines[cursorY].slice(0, cursorX);
          lines.splice(cursorY + 1);
        } else if (mode === 1) {
          lines[cursorY] = ' '.repeat(cursorX) + lines[cursorY].slice(cursorX);
          for (let j = 0; j < cursorY; j++) {
            lines[j] = ' '.repeat(lines[j].length);
          }
        } else if (mode === 2 || mode === 3) {
          lines.fill('');
          cursorX = 0;
          cursorY = 0;
        }
      } else if (command === 'K') { // Erase in Line
        eraseLine(parseInt(sequence || '0'));
      }
    } else if (input[i] === '\n') {
      cursorY++;
      cursorX = 0;
      if (cursorY >= lines.length) {
        lines.push('');
      }
    } else if (input[i] === '\r') {
      cursorX = 0;
    } else {
      if (cursorX >= lines[cursorY].length) {
        lines[cursorY] += ' '.repeat(cursorX - lines[cursorY].length) + input[i];
      } else {
        lines[cursorY] = lines[cursorY].slice(0, cursorX) + input[i] + lines[cursorY].slice(cursorX + 1);
      }
      cursorX++;
    }
  }

  // Remove any empty lines at the end
  while (lines.length > 0 && lines[lines.length - 1].trim() === '') {
    lines.pop();
  }

  return lines.join('\n');
};

const LogLine = ({ message, docker, isMobile }) => {
  let html = decodeUnicode((message))
    .replace('\u0001\u0000\u0000\u0000\u0000\u0000\u0000', '')
    .replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;')
    .replace(/(?:\r\n|\r|\n)/g, '<br>')
    .replace(/�/g, '')
    .replace(/ /g, '&nbsp;')
    .replace(/\t/g, '&nbsp;&nbsp;&nbsp;&nbsp;')
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
    
    return <Stack direction={isMobile ? 'column' : 'row'} spacing={1} alignItems="flex-start">
    <div style={{color:'#AAAAFF', fontStyle:'italic', whiteSpace: 'pre', background: '#393f48', padding: '0 0.5em', marginRight: '5px'}}>
      {parts[0].replace('T', ' ').split('.')[0]}
    </div>
    <div dangerouslySetInnerHTML={{ __html: restString }} />
  </Stack>;
  }
    
  return  <div dangerouslySetInnerHTML={{ __html: html}} />;
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
      return '#5a5af4';
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