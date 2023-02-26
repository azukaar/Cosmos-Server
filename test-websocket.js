const WebSocket = require('ws');

// send 

const ws = new WebSocket('ws://localhost:8080/proxy/ws', { 
  rejectUnauthorized: false,
  followRedirects : true,
 });

ws.on('open', function open() {
  ws.send('something');
});

ws.on('message', function incoming(data) {
  console.log(data);
});
