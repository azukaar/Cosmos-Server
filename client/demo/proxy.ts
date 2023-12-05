import { Request, Response } from "express"
import { resolve } from "path"
import { readFileSync } from "fs"
import jsonServer from "json-server"
import traverse from "traverse"

const PROXY_PORT = process.env.PROXY_PORT || 9000

const definedDB: Record<string, any> = JSON.parse(readFileSync(resolve(__dirname, "db.json"), {
    encoding: "utf8"
}))

const autoDB = traverse(definedDB)
    .paths()
    .map(pathArray => ({
        path: pathArray.join("."),
        data: pathArray.reduce((db, pathSegment) => db[pathSegment], definedDB)
    }))
    .map(dbEntry => (dbEntry.data = isObject(dbEntry.data) ? dbEntry.data : ["$return", dbEntry.data], dbEntry))
    .reduce((acc, dbEntry) => (acc[dbEntry.path] = dbEntry.data, acc), {} as Record<string, any>)
const db = Object.assign({}, autoDB, definedDB)
const dbPaths = Object.keys(db)

// TODO: Reference types do not work at the ["$return", data] level
// A possible solution is to replace filter with map and custom behavior for get, post and put
// defining them to interact with the parent object

const definedRoutes: Record<string, string> = JSON.parse(readFileSync(resolve(__dirname, "routes.json"), {
    encoding: "utf8"
}))

const autoRoutes = dbPaths
    .filter(path => path.includes("."))
    .map(path => ({
        origin: "/" + path.replace(/\./g, "/"),
        destination: "/" + path
    }))
    .reduce((acc, bind) => (acc[bind.origin] = bind.destination, acc), {} as Record<string, string>)
const routes = Object.assign({}, autoRoutes, definedRoutes)

const proxyServer = jsonServer.create()
const middlewares = jsonServer.defaults()
const rewriter = jsonServer.rewriter(routes)
const router = jsonServer.router(db, {
    foreignKeySuffix: "Id"
})

// @ts-ignore
router.render = (req: Request, res: Response) => {
    const reqPath = req.path.replace("/", "")
    const matchedPath = dbPaths
        .map(path => path.replace(/\./g, "/"))
        .find(path => path !== reqPath && reqPath.startsWith(path))
    console.log(matchedPath)

    if (typeof res.locals.data["_comment"] === "string" && res.locals.data["_comment"].toLowerCase().includes("todo"))
        return res.sendStatus(404)
    else if (Array.isArray(res.locals.data.list))
        res.locals.data = res.locals.data.list;
    else if (res.locals.data[0] === "$return")
        res.locals.data = res.locals.data[1]

    switch (req.method) {
        case "GET":
            res.locals.result = res.locals.data
            break
        case "PUT": case "POST": default:
            res.locals.result = req.body
            break
    }

    return res.jsonp({
        data: res.locals.result,
        status: "ok"
    })
}

proxyServer.use(jsonServer.bodyParser)
proxyServer.use(middlewares)

proxyServer.use("/api", rewriter)
proxyServer.use("/api", router)
proxyServer.listen(PROXY_PORT, () => console.info(`JSON proxy is running on ${PROXY_PORT} port`))

function isObject(value: unknown) {
    return typeof value === 'object' && value !== null
}