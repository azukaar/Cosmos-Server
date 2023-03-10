var version = readJsonFile("gupm.json").version

var archi = ""
if(typeof $1 != "undefined") {
    archi = $1
    version = version + "-" + archi
}

console.log("Pushing azukaar/cosmos-server:"+version)

var buildSettings = ["build", "--tag", "azukaar/cosmos-server:"+version]

if(archi == "arm64") {
    buildSettings.push("--platform")
    buildSettings.push("linux/arm/v7")
    buildSettings.push("--file")
    buildSettings.push("./dockerfile.arm64")
    buildSettings.push("--tag")
    buildSettings.push("azukaar/cosmos-server:latest-arm64")
} else {
    buildSettings.push("--tag")
    buildSettings.push("azukaar/cosmos-server:latest")
}

buildSettings.push(".")

console.log(buildSettings)

exec("docker", buildSettings)

exec("docker", ["push", "azukaar/cosmos-server:"+version])

if(archi == "arm64") {
    exec("docker", ["push", "azukaar/cosmos-server:latest-arm64"])    
} else {
    exec("docker", ["push", "azukaar/cosmos-server:latest"])
}