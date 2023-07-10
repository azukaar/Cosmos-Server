
const express = require('express')
const app = express()
const port = 3000

// console log every request sent
app.use((req, res, next) => {
  console.log(`[REQ] - ${req.method} ${req.url}`)
  next()
});

app.get('/return/:status/:time', async (req, res) => {
  const statusCode = parseInt(req.params.status);
  const returnString =`Hello status ${statusCode} after ${req.params.time}ms !`

  console.log(`[RES] - ${statusCode} ${returnString}`)
  
  await new Promise(resolve => setTimeout(resolve, req.params.time));

  return res.status(statusCode).send(returnString)
});

app.get('/', (req, res) => {
  console.log("[RES] - Hello World!")
  res.send('Hello World!')
})

app.listen(port, () => {
  console.log(`Example app listening on port ${port}`)
})

// app.ws('/ws', function(ws, req) {
//   ws.on('message', function(msg) {
//     console.log(msg);
//     ws.send(msg);
//   });
// });