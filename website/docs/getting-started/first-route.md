---
id: first-route
title: First Route
description: Read params, query values, JSON input, and returned errors from one handler.
sidebar_position: 3
---

A Zinc handler receives a request-scoped `*zinc.Context` and returns an error.

```go
func(c *zinc.Context) error
```

That same shape is used for route handlers and middleware.

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

## Bind JSON

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

`Bind().All` reads route params, query values, body data, and validation when a validator is configured.

## Return errors

Handlers return errors instead of writing error responses by hand.

```go
app.Get("/users/{id}", func(c *zinc.Context) error {
	user, err := findUser(c.Param("id"))
	if err != nil {
		return zinc.ErrNotFound.WithMessage("user not found")
	}
	return c.JSON(user)
})
```

Zinc's error handler turns known HTTP errors into responses. You can replace that behavior in [Configuration](../guide/configuration).

## Continue

- [Routing](../guide/routing) covers route shapes, groups, and named routes.
- [Context](../guide/context) covers request and response helpers.
- [Binding](../guide/binding) covers the full binding API.
