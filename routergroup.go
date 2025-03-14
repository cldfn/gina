// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package gin

import (
	"net/http"
	"path"
	"regexp"
	"strings"
)

var (
	// regEnLetter matches english letters for http method name
	regEnLetter = regexp.MustCompile("^[A-Z]+$")

	// anyMethods for RouterGroup Any method
	anyMethods = []string{
		http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch,
		http.MethodHead, http.MethodOptions, http.MethodDelete, http.MethodConnect,
		http.MethodTrace,
	}
)

// IRouter defines all router handle interface includes single and group router.
type IRouter[T any] interface {
	IRoutes[T]
	Group(string, ...OldHandlerFunc[T]) *RouterGroup[T]
}

// IRoutes defines all router handle interface.
type IRoutes[T any] interface {
	Use(...OldHandlerFunc[T]) IRoutes[T]
	UseWithAccess(...HandlerFunc[T]) IRoutes[T]

	Handle(string, string, ...OldHandlerFunc[T]) IRoutes[T]
	HandleWithAccess(string, string, ...HandlerFunc[T]) IRoutes[T]

	Any(string, ...OldHandlerFunc[T]) IRoutes[T]
	GET(string, ...OldHandlerFunc[T]) IRoutes[T]
	POST(string, ...OldHandlerFunc[T]) IRoutes[T]
	DELETE(string, ...OldHandlerFunc[T]) IRoutes[T]
	PATCH(string, ...OldHandlerFunc[T]) IRoutes[T]
	PUT(string, ...OldHandlerFunc[T]) IRoutes[T]
	OPTIONS(string, ...OldHandlerFunc[T]) IRoutes[T]
	HEAD(string, ...OldHandlerFunc[T]) IRoutes[T]
	Match([]string, string, ...OldHandlerFunc[T]) IRoutes[T]

	StaticFile(string, string) IRoutes[T]
	StaticFileFS(string, string, http.FileSystem) IRoutes[T]
	Static(string, string) IRoutes[T]
	StaticFS(string, http.FileSystem) IRoutes[T]
}

// RouterGroup is used internally to configure router, a RouterGroup is associated with
// a prefix and an array of handlers (middleware).
type RouterGroup[T any] struct {
	Handlers HandlersChain[T]
	basePath string
	engine   *Engine[T]
	root     bool
}

var _ IRouter[any] = (*RouterGroup[any])(nil)

// Use adds middleware to the group, see example code in GitHub.
func (group *RouterGroup[T]) Use(middleware ...OldHandlerFunc[T]) IRoutes[T] {
	group.Handlers = append(group.Handlers, wrapOldHandlers(middleware)...)
	return group.returnObj()
}

func (group *RouterGroup[T]) UseWithAccess(middleware ...HandlerFunc[T]) IRoutes[T] {
	group.Handlers = append(group.Handlers, middleware...)
	return group.returnObj()
}

// Group creates a new router group. You should add all the routes that have common middlewares or the same path prefix.
// For example, all the routes that use a common middleware for authorization could be grouped.
func (group *RouterGroup[T]) Group(relativePath string, handlers ...OldHandlerFunc[T]) *RouterGroup[T] {
	return &RouterGroup[T]{
		Handlers: group.oldCombineHandlers(handlers),
		basePath: group.calculateAbsolutePath(relativePath),
		engine:   group.engine,
	}
}

// BasePath returns the base path of router group.
// For example, if v := router.Group("/rest/n/v1/api"), v.BasePath() is "/rest/n/v1/api".
func (group *RouterGroup[T]) BasePath() string {
	return group.basePath
}

func (group *RouterGroup[T]) handle(httpMethod, relativePath string, handlers OldHandlersChain[T]) IRoutes[T] {
	absolutePath := group.calculateAbsolutePath(relativePath)
	newhandlers := group.oldCombineHandlers(handlers)
	group.engine.addRoute(httpMethod, absolutePath, newhandlers)
	return group.returnObj()
}

func (group *RouterGroup[T]) handleNew(httpMethod, relativePath string, handlers HandlersChain[T]) IRoutes[T] {
	absolutePath := group.calculateAbsolutePath(relativePath)
	handlers = group.combineHandlers(handlers)
	group.engine.addRoute(httpMethod, absolutePath, handlers)
	return group.returnObj()
}

// Handle registers a new request handle and middleware with the given path and method.
// The last handler should be the real handler, the other ones should be middleware that can and should be shared among different routes.
// See the example code in GitHub.
//
// For GET, POST, PUT, PATCH and DELETE requests the respective shortcut
// functions can be used.
//
// This function is intended for bulk loading and to allow the usage of less
// frequently used, non-standardized or custom methods (e.g. for internal
// communication with a proxy).
func (group *RouterGroup[T]) Handle(httpMethod, relativePath string, handlers ...OldHandlerFunc[T]) IRoutes[T] {
	if matched := regEnLetter.MatchString(httpMethod); !matched {
		panic("http method " + httpMethod + " is not valid")
	}
	return group.handle(httpMethod, relativePath, handlers)
}

func (group *RouterGroup[T]) HandleWithAccess(httpMethod, relativePath string, handlers ...HandlerFunc[T]) IRoutes[T] {
	if matched := regEnLetter.MatchString(httpMethod); !matched {
		panic("http method " + httpMethod + " is not valid")
	}
	return group.handleNew(httpMethod, relativePath, handlers)
}

// POST is a shortcut for router.Handle("POST", path, handlers).
func (group *RouterGroup[T]) POST(relativePath string, handlers ...OldHandlerFunc[T]) IRoutes[T] {
	return group.handle(http.MethodPost, relativePath, handlers)
}

// GET is a shortcut for router.Handle("GET", path, handlers).
func (group *RouterGroup[T]) GET(relativePath string, handlers ...OldHandlerFunc[T]) IRoutes[T] {
	return group.handle(http.MethodGet, relativePath, handlers)
}

// DELETE is a shortcut for router.Handle("DELETE", path, handlers).
func (group *RouterGroup[T]) DELETE(relativePath string, handlers ...OldHandlerFunc[T]) IRoutes[T] {
	return group.handle(http.MethodDelete, relativePath, handlers)
}

// PATCH is a shortcut for router.Handle("PATCH", path, handlers).
func (group *RouterGroup[T]) PATCH(relativePath string, handlers ...OldHandlerFunc[T]) IRoutes[T] {
	return group.handle(http.MethodPatch, relativePath, handlers)
}

// PUT is a shortcut for router.Handle("PUT", path, handlers).
func (group *RouterGroup[T]) PUT(relativePath string, handlers ...OldHandlerFunc[T]) IRoutes[T] {
	return group.handle(http.MethodPut, relativePath, handlers)
}

// OPTIONS is a shortcut for router.Handle("OPTIONS", path, handlers).
func (group *RouterGroup[T]) OPTIONS(relativePath string, handlers ...OldHandlerFunc[T]) IRoutes[T] {
	return group.handle(http.MethodOptions, relativePath, handlers)
}

// HEAD is a shortcut for router.Handle("HEAD", path, handlers).
func (group *RouterGroup[T]) HEAD(relativePath string, handlers ...OldHandlerFunc[T]) IRoutes[T] {
	return group.handle(http.MethodHead, relativePath, handlers)
}

// Any registers a route that matches all the HTTP methods.
// GET, POST, PUT, PATCH, HEAD, OPTIONS, DELETE, CONNECT, TRACE.
func (group *RouterGroup[T]) Any(relativePath string, handlers ...OldHandlerFunc[T]) IRoutes[T] {
	for _, method := range anyMethods {
		group.handle(method, relativePath, handlers)
	}

	return group.returnObj()
}

// Match registers a route that matches the specified methods that you declared.
func (group *RouterGroup[T]) Match(methods []string, relativePath string, handlers ...OldHandlerFunc[T]) IRoutes[T] {
	for _, method := range methods {
		group.handle(method, relativePath, handlers)
	}

	return group.returnObj()
}

// StaticFile registers a single route in order to serve a single file of the local filesystem.
// router.StaticFile("favicon.ico", "./resources/favicon.ico")
func (group *RouterGroup[T]) StaticFile(relativePath, filepath string) IRoutes[T] {
	return group.staticFileHandler(relativePath, WrapHandler(func(c *Context[T]) {
		c.File(filepath)
	}))
}

// StaticFileFS works just like `StaticFile` but a custom `http.FileSystem` can be used instead..
// router.StaticFileFS("favicon.ico", "./resources/favicon.ico", Dir{".", false})
// Gin by default uses: gin.Dir()
func (group *RouterGroup[T]) StaticFileFS(relativePath, filepath string, fs http.FileSystem) IRoutes[T] {
	return group.staticFileHandler(relativePath, WrapHandler(func(c *Context[T]) {
		c.FileFromFS(filepath, fs)
	}))
}

func (group *RouterGroup[T]) staticFileHandler(relativePath string, handler HandlerFunc[T]) IRoutes[T] {
	if strings.Contains(relativePath, ":") || strings.Contains(relativePath, "*") {
		panic("URL parameters can not be used when serving a static file")
	}

	// todo optimize
	group.handleNew("GET", relativePath, HandlersChain[T]{handler})
	group.handleNew("HEAD", relativePath, HandlersChain[T]{handler})
	return group.returnObj()
}

// Static serves files from the given file system root.
// Internally a http.FileServer is used, therefore http.NotFound is used instead
// of the Router's NotFound handler.
// To use the operating system's file system implementation,
// use :
//
//	router.Static("/static", "/var/www")
func (group *RouterGroup[T]) Static(relativePath, root string) IRoutes[T] {
	return group.StaticFS(relativePath, Dir(root, false))
}

// StaticFS works just like `Static()` but a custom `http.FileSystem` can be used instead.
// Gin by default uses: gin.Dir()
func (group *RouterGroup[T]) StaticFS(relativePath string, fs http.FileSystem) IRoutes[T] {
	if strings.Contains(relativePath, ":") || strings.Contains(relativePath, "*") {
		panic("URL parameters can not be used when serving a static folder")
	}
	handler := group.createStaticHandler(relativePath, fs)
	urlPattern := path.Join(relativePath, "/*filepath")

	// Register GET and HEAD handlers
	group.handleNew("GET", urlPattern, HandlersChain[T]{handler})
	group.handleNew("HEAD", urlPattern, HandlersChain[T]{handler})
	return group.returnObj()
}

func (group *RouterGroup[T]) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc[T] {
	absolutePath := group.calculateAbsolutePath(relativePath)
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))

	return WrapHandler(func(c *Context[T]) {
		if _, noListing := fs.(*OnlyFilesFS); noListing {
			c.Writer.WriteHeader(http.StatusNotFound)
		}

		file := c.Param("filepath")
		// Check if file exists and/or if we have permission to access it
		f, err := fs.Open(file)
		if err != nil {
			c.Writer.WriteHeader(http.StatusNotFound)
			c.handlers = group.engine.noRoute
			// Reset index
			c.index = -1
			return
		}
		f.Close()

		fileServer.ServeHTTP(c.Writer, c.Request)
	})
}

func wrapOldHandlers[T any](handlers OldHandlersChain[T]) HandlersChain[T] {

	result := make([]HandlerFunc[T], len(handlers))

	for idx, it := range handlers {
		result[idx] = WrapHandler(it)
	}

	return result
}

func (group *RouterGroup[T]) oldCombineHandlers(handlers OldHandlersChain[T]) HandlersChain[T] {
	finalSize := len(group.Handlers) + len(handlers)
	assert1(finalSize < int(abortIndex), "too many handlers")
	mergedHandlers := make(HandlersChain[T], finalSize)
	copy(mergedHandlers, group.Handlers)
	copy(mergedHandlers[len(group.Handlers):], wrapOldHandlers(handlers))
	return mergedHandlers
}

func (group *RouterGroup[T]) combineHandlers(handlers HandlersChain[T]) HandlersChain[T] {
	finalSize := len(group.Handlers) + len(handlers)
	assert1(finalSize < int(abortIndex), "too many handlers")
	mergedHandlers := make(HandlersChain[T], finalSize)
	copy(mergedHandlers, group.Handlers)
	copy(mergedHandlers[len(group.Handlers):], handlers)
	return mergedHandlers
}

func (group *RouterGroup[T]) calculateAbsolutePath(relativePath string) string {
	return joinPaths(group.basePath, relativePath)
}

func (group *RouterGroup[T]) returnObj() IRoutes[T] {
	if group.root {
		return group.engine
	}
	return group
}
