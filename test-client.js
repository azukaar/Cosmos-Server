const https = require('https');

const options = {
  hostname: 'radarr.yann-server.com', // Replace with your target URL
  method: 'OPTIONS',
  port: 443,
  path: '/cosmos-ui', // Replace with your target path
};

const req = https.request(options, (res) => {
  console.log('STATUS:', res.statusCode);
  console.log('HEADERS:', JSON.stringify(res.headers, null, 2));

  res.setEncoding('utf8');
  res.on('data', (chunk) => {
    console.log(`BODY: ${chunk}`);
  });
});

req.on('error', (e) => {
  console.error(`problem with request: ${e.message}`);
});

// Close the request
req.end();