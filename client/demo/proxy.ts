import jsonServer from "json-server"

const PROXY_PORT = process.env.PROXY_PORT || 9000

const proxyServer = jsonServer.create()
const router = jsonServer.router("db.json")
const middlewares = jsonServer.defaults()

proxyServer.use(middlewares)
proxyServer.use("/api", router)
proxyServer.listen(PROXY_PORT, () => console.info(`JSON proxy is running on ${PROXY_PORT} port`))