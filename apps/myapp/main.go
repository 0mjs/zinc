package main

import (
	"encoding/json"
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/0mjs/zinc"
)

func main() {
	app := zinc.New()

	// Initialize template engine
	templateEngine := zinc.NewTemplateEngine("./templates")
	templateEngine.AddFunc("formatDate", func(t time.Time) string {
		return t.Format("2006-01-02")
	})
	if err := templateEngine.LoadTemplates(); err != nil {
		log.Printf("Templates directory: %s", templateEngine.GetDirectory())
		log.Fatal("Failed to load templates:", err)
	}
	app.SetTemplateEngine(templateEngine)

	// Initialize WebSocket
	app.SetWebSocketHandler(zinc.WebSocket())

	// Initialize file upload handler
	fileUpload := zinc.NewFileUpload("uploads")
	fileUpload.SetMaxSize(5 << 20) // 5MB
	fileUpload.SetAllowedTypes([]string{".jpg", ".png", ".pdf"})
	app.SetFileUpload(fileUpload)

	// Implicit string
	// Note: This isn't overly useful, but it's a good example of how to implicitly use context.
	app.Get("/", helloWorld)

	// Implicit handler method
	// Note: Handler methods simply receive the context as a parameter and must send a response if passed implicitly.
	app.Get("/hello", hello)

	// Explicit string
	// Note: This is the most common way to send a response.
	app.Get("/alt", func(c *zinc.Context) {
		c.Send(helloWorld)
	})

	// JSON
	// Note: This is a specific method for sending JSON responses.
	app.Get("/json", func(c *zinc.Context) {
		c.JSON(zinc.Map{
			"message": helloWorld,
		})
	})

	// Path parameters
	// Note: Path parameters are parsed from the URL path and can be accessed using the Context.Param method.
	app.Get("/users/:id", func(c *zinc.Context) {
		c.JSON(zinc.Map{
			"message": fmt.Sprintf("the user id is %s", c.Param("id")),
		})
	})

	// Nested path parameters
	// Note: Nested path parameters are parsed from the URL path and can be accessed using the Context.Param method.
	app.Get("/users/:userID/posts/:postID", func(c *zinc.Context) {
		c.JSON(zinc.Map{
			"user": c.Param("userID"),
			"post": c.Param("postID"),
		})
	})

	// Query parameters
	// Note: Query parameters are parsed from the URL query string and can be accessed using the Context.Query method.
	app.Get("/search", func(c *zinc.Context) {
		c.JSON(zinc.Map{
			"search": c.Query("q"),
			"limit":  c.Query("limit"),
		})
	})

	// Route grouping
	// Note: Route groups are used to create a new route group.
	apiGroup := app.Group("/api")

	// Route group method calling/context
	// Note: Instatiated groups can be used to call methods on the context.
	apiGroup.Get("/usernames", func(c *zinc.Context) {
		c.JSON(zinc.Map{
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
	externalAPI.Get("/", func(c *zinc.Context) {
		c.JSON(zinc.Map{
			"version": "v1",
			"users":   []string{"martin", "stephen"},
		})
	})

	// - /v1/your-api
	yourApi := v1Group.Group("/your-api")

	// - /v1/your-api/users
	// Note: Methods can be called at all levels of the group hierarchy.
	v2UsersGroup := yourApi.Group("/users")
	v2UsersGroup.Get("/", func(c *zinc.Context) {
		c.JSON(zinc.Map{
			"version": "v2",
			"users":   []string{"magnus", "jason", "svend"},
		})
	})

	// Standard method routing
	// Note: Method routing is used to handle different HTTP methods.
	// - GET
	app.Get("/get-method", func(c *zinc.Context) {
		c.JSON(zinc.Map{
			"method": c.Method,
		})
	})

	// - POST
	app.Post("/post-method", func(c *zinc.Context) {
		c.JSON(zinc.Map{
			"method": c.Method,
		})
	})

	// - PUT
	app.Put("/put-method", func(c *zinc.Context) {
		c.JSON(zinc.Map{
			"method": c.Method,
		})
	})

	// - DELETE
	app.Delete("/delete-method", func(c *zinc.Context) {
		c.JSON(zinc.Map{
			"method": c.Method,
		})
	})

	// - PATCH
	app.Patch("/patch-method", func(c *zinc.Context) {
		c.JSON(zinc.Map{
			"method": c.Method,
		})
	})

	// - HEAD
	app.Head("/head-method", func(c *zinc.Context) {
		c.JSON(zinc.Map{
			"method": c.Method,
		})
	})

	// - OPTIONS
	app.Options("/options-method", func(c *zinc.Context) {
		c.JSON(zinc.Map{
			"method": c.Method,
		})
	})

	// App-level middleware
	// Note: App-level middleware is used to apply middleware to all routes, regardless of the route group.
	app.Use(MyMiddleware())

	// Route-level middleware
	// Note: Route-level middleware is used to apply middleware to a specific route.
	app.Get("/middleware", MyMiddleware(), func(c *zinc.Context) {
		c.JSON(zinc.Map{
			"message": "Hello, Middleware!",
		})
	})

	// Chained middleware
	// Note: Chained middleware is used to apply multiple middleware to a specific route.
	app.Get(
		"/permissions",
		Authenticate(),
		Authorize("some-permission"),
		func(c *zinc.Context) {
			authenticated := c.Get("authenticated")
			authorized := c.Get("authorized")
			c.JSON(zinc.Map{
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
	app.Get("/static", func(c *zinc.Context) {
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

	// Cron jobs
	// Note: Cron jobs are used to run jobs at specified intervals.
	app.Cron("job", "@every 30s", func() {
		fmt.Println("Job executed at", time.Now())
	})

	// Basic route with validation
	app.Post("/users", func(c *zinc.Context) {
		type User struct {
			Name     string `json:"name" validate:"required"`
			Email    string `json:"email" validate:"required,email"`
			Age      int    `json:"age" validate:"min=18,max=100"`
			Password string `json:"password" validate:"required,min=8"`
		}

		var user User
		if err := c.Body(&user); err != nil {
			c.Status(400).JSON(zinc.Map{"error": err.Error()})
			return
		}

		if errs := app.Validate(user); len(errs) > 0 {
			c.Status(400).JSON(zinc.Map{"errors": errs})
			return
		}

		c.JSON(zinc.Map{
			"message": "User validated successfully",
			"user":    user,
		})
	})

	// Template rendering example
	app.Get("/profile", func(c *zinc.Context) {
		profileData := zinc.Map{
			"Name":     "John Doe",
			"Age":      30,
			"JoinDate": time.Now(),
		}
		if err := c.Render("profile", profileData); err != nil {
			c.Status(500).Send(err.Error())
		}
	})

	// Chat room page
	app.Get("/chat/:room", func(c *zinc.Context) {
		data := zinc.Map{
			"RoomID":   c.Param("room"),
			"RandomID": fmt.Sprintf("%d", time.Now().UnixNano()%10000),
		}
		if err := c.Render("chat", data); err != nil {
			c.Status(500).Send(err.Error())
		}
	})

	// WebSocket chat endpoint
	app.Get("/ws/chat/:room", func(c *zinc.Context) {
		roomID := c.Param("room")

		// Upgrade HTTP connection to WebSocket
		conn, err := c.Upgrade()
		if err != nil {
			c.Status(400).Send(err.Error())
			return
		}

		// Join the room
		c.JoinRoom(roomID, conn)
		defer c.LeaveRoom(roomID, conn)

		for {
			// Read message from WebSocket
			_, rawMsg, err := conn.ReadMessage()
			if err != nil {
				break
			}

			// Parse the message
			var msg map[string]interface{}
			if err := json.Unmarshal(rawMsg, &msg); err != nil {
				continue
			}

			// Handle ping messages
			if msgType, ok := msg["type"].(string); ok && msgType == "ping" {
				continue
			}

			// Broadcast the message to all clients in the room
			if err := c.BroadcastToRoom(roomID, rawMsg); err != nil {
				break
			}
		}
	})

	// File upload example
	app.Post("/upload", func(c *zinc.Context) {
		file, err := c.File("file")
		if err != nil {
			c.Status(400).JSON(zinc.Map{"error": err.Error()})
			return
		}

		filename, err := c.SaveFile(file)
		if err != nil {
			c.Status(500).JSON(zinc.Map{"error": err.Error()})
			return
		}

		c.JSON(zinc.Map{
			"message":  "File uploaded successfully",
			"filename": filename,
		})
	})

	// Multiple file upload example
	app.Post("/upload/multiple", func(c *zinc.Context) {
		files, err := c.Files("files")
		if err != nil {
			c.Status(400).JSON(zinc.Map{"error": err.Error()})
			return
		}

		var uploaded []string
		for _, file := range files {
			filename, err := c.SaveFile(file)
			if err != nil {
				c.Status(500).JSON(zinc.Map{"error": err.Error()})
				return
			}
			uploaded = append(uploaded, filename)
		}

		c.JSON(zinc.Map{
			"message": "Files uploaded successfully",
			"files":   uploaded,
		})
	})

	// Example of using all features together
	app.Post("/posts", func(c *zinc.Context) {
		// Validate post data
		type Post struct {
			Title   string `json:"title" validate:"required,min=3"`
			Content string `json:"content" validate:"required,min=10"`
		}

		var post Post
		if err := c.Body(&post); err != nil {
			c.Status(400).JSON(zinc.Map{"error": err.Error()})
			return
		}

		if errs := app.Validate(post); len(errs) > 0 {
			c.Status(400).JSON(zinc.Map{"errors": errs})
			return
		}

		// Handle file upload
		file, err := c.File("image")
		if err != nil {
			c.Status(400).JSON(zinc.Map{"error": "Image is required"})
			return
		}

		filename, err := c.SaveFile(file)
		if err != nil {
			c.Status(500).JSON(zinc.Map{"error": err.Error()})
			return
		}

		// Render response using template
		data := zinc.Map{
			"Title":   post.Title,
			"Content": post.Content,
			"Image":   filename,
			"Created": time.Now(),
		}

		if err := c.Render("post", data); err != nil {
			c.Status(500).Send(err.Error())
			return
		}

		// Broadcast new post to WebSocket subscribers
		msg := fmt.Sprintf("New post: %s", post.Title)
		c.BroadcastToRoom("posts", []byte(msg))
	})

	// Start the server
	log.Fatal(app.Serve(":8080"))
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
