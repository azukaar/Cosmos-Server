import { Stack } from '@mui/material';
import DOMPurify from 'dompurify';

function decodeUnicode(str) {
  return str.replace(/\\u([0-9a-fA-F]{4})/g, (match, p1) => {
    return String.fromCharCode(parseInt(p1, 16));
  });
}

const LogLine = ({ message, docker, isMobile }) => {
  let html = decodeUnicode(message)
    .replace('\u0001\u0000\u0000\u0000\u0000\u0000\u0000', '')
    .replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;')
    .replace(/(?:\r\n|\r|\n)/g, '<br>')
    .replace(/�/g, '')
    .replace(/ /g, '&nbsp;')
    .replace(/\t/g, '&nbsp;&nbsp;&nbsp;&nbsp;');

  // Process color codes
  let colorStack = [];
  html = html.replace(/\x1b\[([0-9]{1,2}(?:;[0-9]{1,2})*)?m/g, (match, p1) => {
    console.log(p1)
    if (!p1 || p1 === '0') {
      // Reset code
      let i = colorStack.length;
      colorStack = [];
      return '</span>'.repeat(i);
    }
    const codes = p1.split(';');
    const styles = [];
    for (const code of codes) {
      const style = getColor(code);
      if (style.startsWith('background-color') || style.startsWith('font-')) {
        styles.push(style);
      } else {
        styles.push(`color:${style}`);
      }
    }
    colorStack.push(styles);
    return `<span style="${styles.join(';')}">`;
  });

  // Close any remaining open spans
  html += '</span>'.repeat(colorStack.length);

  if (docker) {
    // Docker prefixes every log line with a UTC RFC3339 timestamp (e.g. 2026-09-07T08:07:46.123456789Z),
    // followed by a space, then the container's own stdout/stderr. Some containers re-print their own
    // local timestamp right after that prefix (e.g. "2026-09-07 10:07:45.720 CEST ..."). The timestamp
    // regex is anchored to the start of the line so we always extract the Docker UTC prefix, and any
    // timestamp the container printed itself is removed from the body to avoid duplicating it.
    let parts = html.match(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+Z/);
    if (!parts) {
      return <div dangerouslySetInnerHTML={{ __html: DOMPurify.sanitize(html) }} />;
    }
    // Convert the Docker UTC timestamp to the viewer's local timezone for display.
    let ts = parts[0].replace(/\.(\d{3})\d+Z/, '.$1Z'); // JS Date only handles ms precision
    let localDate = new Date(ts);
    const pad2 = (n) => String(n).padStart(2, '0');
    const formatted = isNaN(localDate.getTime())
      ? parts[0].replace('T', ' ').split('.')[0]  // fall back to raw UTC
      : `${localDate.getFullYear()}-${pad2(localDate.getMonth() + 1)}-${pad2(localDate.getDate())} ${pad2(localDate.getHours())}:${pad2(localDate.getMinutes())}:${pad2(localDate.getSeconds())}`;

    // Remove the Docker prefix, then any timestamp the container printed itself right after it
    // (e.g. "2026-09-07 10:07:45.720 CEST" or "... +02:00") so it does not appear twice.
    let restString = html.replace(parts[0], '')
      .replace(/^&nbsp;/, '')
      .replace(/^(?:<span[^>]*>)*\[?\d{4}(?:-|\/)\d{2}(?:-|\/)\d{2}(?:T|&nbsp;)\d{2}:\d{2}:\d{2}(?:[.,]\d{1,9})?(?:Z|[+-]\d{2}:?\d{2})?(?:&nbsp;(?:[A-Z]{2,6}|[+-]\d{2}:?\d{2}|Z))?\]?(?:<\/span>)*(?:&nbsp;+|<br>|$)/, '')
      .replace(/^&nbsp;(?=(?:&nbsp;)*[A-Z\[])/, '');

    return (
      <Stack direction={isMobile ? 'column' : 'row'} spacing={1} alignItems="flex-start">
        <div style={{color:'#AAAAFF', fontStyle:'italic', whiteSpace: 'pre', background: '#393f48', padding: '0 0.5em', marginRight: '5px'}}>
          {formatted}
        </div>
        <div dangerouslySetInnerHTML={{ __html: DOMPurify.sanitize(restString) }} />
      </Stack>
    );
  }
   
  return <div dangerouslySetInnerHTML={{ __html: DOMPurify.sanitize(html) }} />;
};

const getColor = (code) => {
  switch (code) {
    // Bold text
    case '1': return 'font-weight: bold';
    // Normal text colors
    case '30': return 'black';
    case '31': return 'red';
    case '32': return 'green';
    case '33': return 'yellow';
    case '34': return '#5a5af4';
    case '35': return 'magenta';
    case '36': return 'cyan';
    case '37': return 'gray';
    case '97': return 'white';
    // Bright text colors
    case '90': return 'darkgray';
    case '91': return 'lightred';
    case '92': return 'lightgreen';
    case '93': return 'lightyellow';
    case '94': return 'lightblue';
    case '95': return 'lightmagenta';
    case '96': return 'lightcyan';
    // Background colors
    case '40': return 'background-color: black';
    case '41': return 'background-color: red';
    case '42': return 'background-color: green';
    case '43': return 'background-color: yellow';
    case '44': return 'background-color: #5a5af4';
    case '45': return 'background-color: magenta';
    case '46': return 'background-color: cyan';
    case '47': return 'background-color: gray';
    case '107': return 'background-color: white';
    default: return 'inherit';
  }
};

export default LogLine;