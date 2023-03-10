// look up dependencies in local go_modules and gupm_modules directories
env("GOPATH", run("go", ["env", "GOROOT"]) + ":" + pwd() + "/go_modules" + ":" + pwd() + "/gupm_modules")
env("GO111MODULE", "off")
env("LOG_LEVEL", "DEBUG")
// dev mode
env("MONGODB", readJsonFile("dev.json").MONGODB)
env("HTTP_PORT", 8080)
env("HTTPS_PORT", 8443)
env("CONFIG_FILE", "./config_dev.json")
env("EZ", "UTC")