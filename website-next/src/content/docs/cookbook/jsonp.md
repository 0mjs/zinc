---
title: JSONP
description: Return a carefully validated JSONP response for a legacy integration.
---

JSONP executes JavaScript supplied by another origin. Prefer CORS and normal JSON for new applications.

```go
var callbackName = regexp.MustCompile(`^[A-Za-z_$][A-Za-z0-9_$\.]{0,127}$`)

app.Get("/legacy", func(c *zinc.Context) error {
    callback := c.Query("callback")
    if !callbackName.MatchString(callback) {
        return zinc.NewError(zinc.StatusBadRequest).WithMessage("invalid callback")
    }

    payload, err := json.Marshal(zinc.Map{"ok": true})
    if err != nil {
        return err
    }

    body := callback + "(" + string(payload) + ");"
    return c.Data("application/javascript; charset=utf-8", []byte(body))
})
```

Never place an unvalidated callback value into a JavaScript response.
