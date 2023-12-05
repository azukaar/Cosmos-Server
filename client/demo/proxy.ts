import { Request, Response } from "express"
import { resolve } from "path"
import { readFileSync } from "fs"
import { stringify } from "querystring"
import jsonServer from "json-server"
import traverse from "traverse"

const PROXY_PORT = process.env.PROXY_PORT || 9000

const definedDB: Record<string, any> = JSON.parse(readFileSync(resolve(__dirname, "db.json"), {
    encoding: "utf8"
}))

const autoDB = genDB(definedDB)
const dbPaths = Object.keys(autoDB)

console.log(dbPaths)

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
const router = jsonServer.router(autoDB, {
    foreignKeySuffix: "Id"
})

// @ts-ignore
router.render = (req: Request, res: Response) => {
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
proxyServer.use("/api", (req, res, next) => {
    const query = stringify(req.query as Record<string, any>)
    const dbKey = req.path
        .replace("/", "")
        .replace(/\//g, ".")

    if (autoDB[dbKey])
        req.url = `/${dbKey}${query ? "?" : ""}${query}`
    next()
})

proxyServer.use("/api", router)
proxyServer.listen(PROXY_PORT, () => console.info(`JSON proxy is running on ${PROXY_PORT} port`))

type JSONValue = string | number | boolean | JSONObject | JSONArray

interface JSONObject extends Record<string, JSONValue> { }
interface JSONArray extends Array<JSONValue> { }

function genDB(initialDB: JSONObject, skipRoot: boolean = true): Record<string, JSONValue> {
    return traverse(initialDB)
        .reduce(function (acc: Record<string, any>, node) {
            if (skipRoot && this.isRoot)
                return acc
            const currentPath = this.path.join(".")

            acc[currentPath] = isObject(node) ? node : ["$return", node]

            if (Array.isArray(node)) {
                node
                    .filter(isObject)
                    .map(childNode => [
                        Object.values(childNode),
                        genDB(childNode, false)
                    ] as const)
                    .forEach(([values, childNodeDB]) =>
                        values
                            .filter(value => !!value && typeof value === "string" && !value.includes("."))
                            .forEach(value => Object.keys(childNodeDB).forEach(childNodePath =>
                                acc[`${currentPath}.${value}${childNodePath && "." || ""
                                }${childNodePath}`.replace(/\//g, ".")] = childNodeDB[childNodePath]
                            ))
                    )
            }

            return acc
        }, {})
}

function isObject(value: unknown): value is Record<string, any> {
    return typeof value === 'object' && value !== null
}

// Проитерировать ключи обьекта
// сделать из них путь к базовому обьекту