# Contribute

Thanks for your interest in contributing to Cosmos, and its source code!

## Getting Started

### Basics

You need Node/NPM, Docker and golang in order to run Cosmos.
Docker can be any version.

Install Docker

```
curl -fsSL https://get.docker.com | sudo sh
```

Install Node **V16**: https://nodejs.org/en/download

Install GOlang v1.21.6: https://go.dev/doc/install

### Required files

Aside from those, you will need a bunch of files. First, You optionally need GeoLite2-Country.tar.gz if you want to test the geolocation (contains IP infos). You can get it from https://maxmind.com/
Second you need the Nebula binaries from https://github.com/slackhq/nebula/releases/download/v1.7.2/nebula-linux-arm64.tar.gz and https://github.com/slackhq/nebula/releases/download/v1.7.2/nebula-linux-amd64.tar.gz
Make sure you have the binary set in the root of your workspace like  this:

![image](https://github.com/azukaar/Cosmos-Server/assets/7872597/11de2778-e799-47b7-b0ba-443e658965dd)


### Setting up and running the dev client

The Front-end is compiled using Vite, a nodejs app that bundles frontend. It is written in React.
First, install dependencies with `npm install`.
Once node is installed, you can simply start the frontend's code with: `npm run client`. This will start the client dev server on :5173

### Building the Client

One you are done developing, you can build the client locally with `npm run client-build`

### Setting up the dev server

Server is a bit more tricky. Once Go is installed, first use `npm run build` to build the server. it will be built in the build/ folder.
Make sure you build the client at least once (previous section) as the build script simply copy paste the output of the `client-build` script.

The server can be started with `npm run start` this will run Cosmos with a bunch of arguments. The most importants are

 * COSMOS_CONFIG_FOLDER=./zz_test_config : this overwrites the config file to be zz_test_config in the cosmos folder (easier to reach)
 * CONFIG_FILE=./config_dev.json : this sets the config file to be config_dev.json

This means that you can find both config files  and config folders at the root of your workspace when running the test build.

When running the client is `npm run client` it will connect to this dev server on port :8443

### Recommandations

It is recommended that you run the local server with 
 * HTTP self-signed: it's better to test in HTTPS than HTTP for closer-to-real-life setup
 * External DB, so you dont lose it during testing stuff (you can use Mongo Atlas for free)
 * Docker on
 * DEBUG level logs
 * Monitoring off for less relevant logs to be removed

### Build and test docker image

You can build a test docker image with `npm run dockerdevbuild` (or `npm run dockerdev` to build it and run it as well). This will use the non-production dockerfile at `dockerfile.local` as opposed to the prod `dockerfile`

### About the demo

There is a demo system integrated that builds a FE with mocked API calls. You can test it with `npm run devdemo` or build it with `npm run demo`

### Other considerations

Try to communicate with the rest of the team **before** you make changes to make sure that it is inline with the vision, and you dont end up doing something we can't merge in. Keep the PR simple, break them into multiple PRs if required.

Thanks again for your consideration, see you soon in the PR section ;)
