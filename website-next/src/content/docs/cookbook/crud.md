---
title: CRUD
description: Structure create, read, update, and delete routes as a Zinc resource group.
---

Group resource operations under one prefix:

```go
widgets := app.Group("/widgets")
widgets.Post("/", createWidget)
widgets.Get("/", listWidgets)
widgets.Get("/{id}", getWidget)
widgets.Put("/{id}", updateWidget)
widgets.Delete("/{id}", deleteWidget)
```

Bind typed request input in create and update handlers:

```go
type WidgetInput struct {
    Name string `json:"name" validate:"required"`
}

func createWidget(c *zinc.Context) error {
    var input WidgetInput
    if err := c.Bind().JSON(&input); err != nil {
        return zinc.NewError(zinc.StatusBadRequest).WithCause(err)
    }

    widget, err := store.Create(c.Context(), input)
    if err != nil {
        return err
    }
    return c.Status(zinc.StatusCreated).JSON(widget)
}
```

The complete persistent example, including schema creation and parameterised SQL, is in [SQLite CRUD API](/cookbook/sqlite-crud-api/).
