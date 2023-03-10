removeFiles("build")

var goArgs = ["build", "-o"]

if(typeof $1 != "undefined" && $1 == "windows") {
    goArgs.push("build/cosmos.exe")
} else {
    goArgs.push("build/cosmos")
}

goArgs = goArgs.concat(dir("src/*.go"))

var archi = "amd64"
if(typeof $2 != "undefined") {
    archi = $2
}

if(typeof $1 != "undefined" && $1 == "mac") {
    env("GOOS", "darwin")
    env("GOARCH", archi)
    exec("go", goArgs)
}
if(typeof $1 != "undefined" && $1 == "windows") {
    env("GOOS", "windows")
    env("GOARCH", archi)
    exec("go", goArgs)
} else {
    env("GOOS", "linux")
    env("GOARCH", archi)
    exec("go", goArgs)
}
