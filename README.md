# KGP - Web Crawler, Indexer, and Search tool
A RESTful API to crawl the web, index it for word count and then search it.

This is a small API to learn Go. It uses **gorilla/mux** for the API.

## Author
Name: Andrew Kunzel

Website: [www.dl3consulting.com](http://www.dl3consulting.com)

## Installation & Run
```bash
# Download this project
go get github.com/kunzel-andrew/kgp
```

Before running API server, you should check the default config values in [config.json](https://github.com/kunzel-andrew/kgp/blob/master/config.json)

```bash
# Build and Run
cd kgp
go build
./kgp

# API Endpoint : http://127.0.0.1:8080
```

## Structure
```
├── main.go             //Creates API Server and Routes
│── handlers.go         //Handlers for the API Routes
│-- indexFuncs.go       //Functions that the index handler uses
|-- searchFuncs.go      //Functions that the search handler uses
│-- config.json         //Configuration File

```

## API

#### /index
* `POST` : Index a Page
    * Takes a JSON Body with the URL to start indexing as a parameter. 
    * Returns a 422 if no URL is found in body
* `DELETE`: Delete the Current Index Cache in Memory

#### /search/:word
* `GET` : Search the Index Cache For A Given Word


### Todo
- [ ] Increase Test Coverage and Test Cases
- [ ] Swagger Documentation
- [ ] Persistent Cache
- [ ] Dockerfile to generate Image
- [ ] Security and Rate Limiting