[![Build Status](https://travis-ci.org/bobinette/papernet.svg?branch=master)](https://travis-ci.org/bobinette/papernet)

# Papernet

Papernet is a very simple tool to keep track of the papers you read.

# Setup

## Go
Usual go installation (1.7.5)

I use the gvm, but you can install golang any way you prefer
```bash
bash < <(curl -s -S -L https://raw.githubusercontent.com/moovweb/gvm/master/binscripts/gvm-installer)
gvm install go1.7.5 --binary
```

then add the following to your bashrc (or the equivalent for your terminal)
```bash
source $HOME/.gvm/scripts/gvm
gvm use go1.7.5
export GOPATH=/path/to/go
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin
```

## Dependencies
The dependencies are handled with `godep`:
```bash
go get -u github.com/tools/godep
godep restore
```

## Web server
To start the web server, you have to create the `data` folder, and setup the index:
```bash
mkdir data
go run cmd/cli/*.go index create --index=data/papernet.index --mapping=bleve/mapping.json
```

Now that everything is ready, you can start the server:
```bash
go run cmd/web/main.go
```

## Front-end
This repository contains the backend of the Papernet project, for the front-end check out https://github.com/bobinette/papernet-front
