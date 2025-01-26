package main

import (
	"fmt"
	"slices"
	"time"

	"github.com/0mjs/zinc"
	z "github.com/0mjs/zinc"
)

// An entire feature overview of everything that `zinc` can do.
func main() {
	app := z.New()

	// - GET parameters & context examples

	// Implicit string
	// Note: This isn't overly useful, but it's a good example of how to implicitly use context.
	app.Get("/", helloWorld)

	// Implicit handler method
	// Note: Handler methods simply receive the context as a parameter and must send a response if passed implicitly.
	app.Get("/hello", hello)

	// Explicit string
	// Note: This is the most common way to send a response.
	app.Get("/alt", func(c *z.Context) {
		c.Send(helloWorld)
	})

	// JSON
	// Note: This is a specific method for sending JSON responses.
	app.Get("/json", func(c *z.Context) {
		c.JSON(z.Map{
			"message": helloWorld,
		})
	})

	// Path parameters
	// Note: Path parameters are parsed from the URL path and can be accessed using the Context.Param method.
	app.Get("/users/:id", func(c *z.Context) {
		c.JSON(z.Map{
			"message": fmt.Sprintf("the user id is %s", c.Param("id")),
		})
	})

	// Nested path parameters
	// Note: Nested path parameters are parsed from the URL path and can be accessed using the Context.Param method.
	app.Get("/users/:userID/posts/:postID", func(c *z.Context) {
		c.JSON(z.Map{
			"user": c.Param("userID"),
			"post": c.Param("postID"),
		})
	})

	// Query parameters
	// Note: Query parameters are parsed from the URL query string and can be accessed using the Context.Query method.
	app.Get("/search", func(c *z.Context) {
		c.JSON(z.Map{
			"search": c.Query("q"),
			"limit":  c.Query("limit"),
		})
	})

	// Route grouping
	// Note: Route groups are used to create a new route group.
	apiGroup := app.Group("/api")

	// Route group method calling/context
	// Note: Instatiated groups can be used to call methods on the context.
	apiGroup.Get("/usernames", func(c *z.Context) {
		c.JSON(z.Map{
			"usernames": []string{"matt", "steven"},
		})
	})

	// Multiple nested route grouping
	// An API that needs the subdomain "v1/" to be used for an external API, that isn't part of the main app.
	// - Example: /v1/your-api/users
	v1Group := apiGroup.Group("/v1")

	// - /v1/external-api
	externalAPI := v1Group.Group("/external-api")
	// Note: Methods can be called on the group, and the context will be passed to the handler method.
	externalAPI.Get("/", func(c *z.Context) {
		c.JSON(z.Map{
			"version": "v1",
			"users":   []string{"martin", "stephen"},
		})
	})

	// - /v1/your-api
	yourApi := v1Group.Group("/your-api")

	// - /v1/your-api/users
	// Note: Methods can be called at all levels of the group hierarchy.
	v2UsersGroup := yourApi.Group("/users")
	v2UsersGroup.Get("/", func(c *z.Context) {
		c.JSON(z.Map{
			"version": "v2",
			"users":   []string{"magnus", "jason", "svend"},
		})
	})

	// Standard method routing
	// Note: Method routing is used to handle different HTTP methods.
	// - GET
	app.Get("/get-method", func(c *z.Context) {
		c.JSON(z.Map{
			"method": c.Method,
		})
	})

	// - POST
	app.Post("/post-method", func(c *z.Context) {
		c.JSON(z.Map{
			"method": c.Method,
		})
	})

	// - PUT
	app.Put("/put-method", func(c *z.Context) {
		c.JSON(z.Map{
			"method": c.Method,
		})
	})

	// - DELETE
	app.Delete("/delete-method", func(c *z.Context) {
		c.JSON(z.Map{
			"method": c.Method,
		})
	})

	// - PATCH
	app.Patch("/patch-method", func(c *z.Context) {
		c.JSON(z.Map{
			"method": c.Method,
		})
	})

	// - HEAD
	app.Head("/head-method", func(c *z.Context) {
		c.JSON(z.Map{
			"method": c.Method,
		})
	})

	// - OPTIONS
	app.Options("/options-method", func(c *z.Context) {
		c.JSON(z.Map{
			"method": c.Method,
		})
	})

	// App-level middleware
	// Note: App-level middleware is used to apply middleware to all routes, regardless of the route group.
	app.Use(MyMiddleware())

	// Route-level middleware
	// Note: Route-level middleware is used to apply middleware to a specific route.
	app.Get("/middleware", MyMiddleware(), func(c *z.Context) {
		c.JSON(z.Map{
			"message": "Hello, Middleware!",
		})
	})

	// Chained middleware
	// Note: Chained middleware is used to apply multiple middleware to a specific route.
	app.Get(
		"/permissions",
		Authenticate(),
		Authorize("some-permission"),
		func(c *z.Context) {
			authenticated := c.Get("authenticated")
			authorized := c.Get("authorized")
			c.JSON(z.Map{
				"message":       "Hello, World!",
				"authenticated": authenticated,
				"authorized":    authorized,
			})
		},
	)

	// Custom HTML
	// Note: This is a templating method for sending HTML responses.
	app.Get("/html", CustomHTML)

	// Static file serving
	// Note: This is a method for serving static files from disk.
	app.Get("/static", func(c *z.Context) {
		c.Static("eg/static/index.html")
	})

	app.Post("/", func(c *zinc.Context) {
		type User struct {
			Name  string `json:"name" validate:"required"`
			Email string `json:"email" validate:"required,email"`
			Age   int    `json:"age" validate:"gte=0,lte=130"`
		}

		var user User
		if err := c.Body(&user); err != nil {
			c.JSON(zinc.Map{
				"error": err.Error(),
			})
			return
		}
		c.Status(201).JSON(user)
	})

	app.Serve()
}

// Helpers

const (
	helloWorld = "Hello, World!"
)

// Note: This is a handler method that sends a response, taking the context as a parameter.
func hello(c *zinc.Context) {
	c.Send(helloWorld)
}

func Authenticate() zinc.Middleware {
	return func(c *zinc.Context) {
		c.Set("authenticated", true)
		c.Next()
	}
}

func Authorize(permission string) zinc.Middleware {
	return func(c *zinc.Context) {
		validPermissions := []string{"some-permission", "another-permission"}

		if !slices.Contains(validPermissions, permission) {
			c.Status(403).JSON(zinc.Map{"error": "Unauthorized"})
			return
		}

		c.Set("authorized", permission)
		c.Next()
	}
}

func CustomHTML(c *zinc.Context) {
	html := `<!DOCTYPE html><html lang="en"><head> <meta charset="UTF-8"> <meta name="viewport" content="width=device-width, initial-scale=1.0"> <title>Hello World</title> <style> body { margin: 0; height: 100vh; display: flex; align-items: center; justify-content: center; background: linear-gradient(135deg, #6366f1, #a855f7); font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Oxygen, Ubuntu, sans-serif; } .container { background: rgba(255, 255, 255, 0.95); padding: 2rem 3rem; border-radius: 1rem; box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04); text-align: center; } h1 { color: #1f2937; margin: 0; font-size: 2.5rem; font-weight: 700; } p { color: #4b5563; margin-top: 1rem; font-size: 1.1rem; } </style></head><body> <div class="container"> <h1>Hello World!</h1> <p>Welcome to your styled endpoint</p> </div></body></html>`
	c.HTML(html)
}

func MyMiddleware() zinc.Middleware {
	return func(c *zinc.Context) {
		fmt.Println("Request received:", c.Request.URL.Path, "Method:", c.Request.Method, "Time:", time.Now().Format(time.RFC3339))
		c.Next()
	}
}
