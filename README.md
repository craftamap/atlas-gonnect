# atlas-gonnect : An Atlassian Connect Framework written in Golang

## Overview [![GoDoc](https://godoc.org/github.com/craftamap/atlas-gonnect?status.svg)](https://godoc.org/github.com/craftamap/atlas-gonnect) [![Go Report Card](https://goreportcard.com/badge/github.com/craftamap/atlas-gonnect)](https://goreportcard.com/report/github.com/craftamap/atlas-gonnect) [![Coverage Status](https://coveralls.io/repos/github/craftamap/atlas-gonnect/badge.svg?branch=master)](https://coveralls.io/github/craftamap/atlas-gonnect?branch=master)

Atlas-Gonnect is an Atlassian Connect Framework written in Golang, inspired by the [Atlassian Connect Spring Boot Framework](https://bitbucket.org/atlassian/atlassian-connect-spring-boot/src/master/)  as well as the [Atlassian Connect Express Framework](https://bitbucket.org/atlassian/atlassian-connect-express/src/master/). The implementation of Atlas-Gonnect is heavily inspired by the letter one.

This project is not associated with Atlassian.

## Install

```
go get github.com/craftamap/atlas-gonnect
```

## Example

`atlas-gonnect` currently uses/depends on `https://github.com/gorilla/mux`,
which serves as router for our webserver.

A minimal example looks like this:

```go
	// We need the config as well as the atlassian-connect-descriptor as an io.Reader
	config, err := os.Open("config.json")
	if err != nil {
		panic(err)
	}
	descriptor, err := os.Open("atlassian-connect.json")
	if err != nil {
		panic(err)
	}

	// Now, we can initialize the addon.
	addon, err := gonnect.NewAddon(config, descriptor)

	// We create a new mux-router
	router := mux.NewRouter()

	// And insert the default request middleware - this middleware should be used on
	// every request/route going to the server
	router.Use(middleware.NewRequestMiddleware(addon, make(map[string]string)))

	// Register the default routes of atlas-gonnect; note that routes is the package
	// github.com/craftamap/atlas-gonnect/routes
	routes.RegisterRoutes(addon, router)

	// To register an route secured by JWT Authentification, use the AuthentificationMiddleware
	router.Handle(
		"/hello-world",
		// note that middleware is the package github.com/craftamap/atlas-gonnect/middleware
		middleware.NewAuthenticationMiddleware(addon, false)(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Do stuff here
			}),
		),
	)
	http.ListenAndServe(":"+strconv.Itoa(addon.Config.Port), router)
```


For a full example, see my [atlas-gonnect-confluence-example](https://github.com/craftamap/atlas-gonnect-confluence-example).

### request context

The request middleware of gonnect is using `context` to expose some important values to the programmer.

The following keys are set (with the following types):

| key               	| description                                               	| type   	|
|-------------------	|-----------------------------------------------------      	|--------	|
| title             	| Name of the Addon                                         	| string 	|
| addonKey          	| Key of the Addon                                          	| string 	|
| localBaseUrl      	| The BaseURL configured in the configuration               	| string 	|
| hostBaseUrl       	| The URL of the Confluence Instance                        	| string 	|
| hostUrl           	| hostUrl is the same as hostBaseUrl(use hostBaseUrl) 	        | string 	|
| hostStylesheetUrl 	| gets the StylesheetUrl of the Confluence/Jira Instance   	| string 	|
| hostScriptUrl     	| The Atlassian Connect JavaScript                      	| string 	|

If the authentication-middleware was used (and successful), the following keys are set:

| key           	| description                                                                         	| type                     	|
|---------------	|-------------------------------------------------------------------------------------	|--------------------------	|
| userAccountId 	| The userAccountId                                                                   	| string                   	|
| clientKey     	| the clientKey of the atlassian tenant / host                                        	| string                   	|
| httpClient    	| A HostRequestHttpClient used for request to the confluence/jira instance; see below 	| *hostrequest.HostRequest 	|



## Author

Fabian Siegel

## License

Apache 2.0.
