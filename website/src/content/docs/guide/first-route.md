---
title: Your First Route
description: Read params, query values, JSON input, and returned errors from Zinc handlers.
---

A Zinc handler receives a request-scoped `*zinc.Context` and returns an error:

```go
func(c *zinc.Context) error
```

Routes and middleware share that shape. Route registration itself stays terse:

```go
app.Get("/users/{id}", showUser)
```

## Read request data

```go
app.Get("/teams/{teamID}/users/{userID}", func(c *zinc.Context) error {
	return c.JSON(zinc.Map{
		"team_id": c.Param("teamID"),
		"user_id": c.Param("userID"),
		"verbose": c.QueryOr("verbose", "false"),
	})
})
```

Route params come from `{name}` segments. Query values come from the request URL.

## Bind JSON and path values

```go
type CreateUserInput struct {
	TeamID int    `path:"teamID"`
	Name   string `json:"name"`
	Email  string `json:"email"`
}

app.Post("/teams/{teamID}/users", func(c *zinc.Context) error {
	var input CreateUserInput
	if err := c.Bind().All(&input); err != nil {
		return err
	}

	return c.Status(zinc.StatusCreated).JSON(input)
})
```

`Bind().All` reads route params, query values, body data, and validation when a
validator is configured.

## Return errors

```go
app.Get("/users/{id}", func(c *zinc.Context) error {
	user, err := findUser(c.Param("id"))
	if err != nil {
		return zinc.ErrNotFound.WithMessage("user not found")
	}
	return c.JSON(user)
})
```

Handlers return errors so Zinc can apply one consistent error policy. Invalid
source-defined route declarations panic during startup; you do not check an
error after every `app.Get` or `app.Post` call.

## Continue

- [Routing](/guide/routing/) covers patterns, groups, and named routes.
- [Binding](/guide/binding/) covers the complete typed binding API.
- [Errors](/guide/errors/) covers HTTP errors and custom error handlers.
