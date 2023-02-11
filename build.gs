removeFiles(["build"]);
var goArgs = ["build", "-o", "build/bin"]
goArgs = goArgs.concat(dir("src/*.go"))
exec("go", goArgs);
copyFiles("client/dist/", "build/static/")