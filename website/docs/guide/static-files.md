---
id: static-files
title: Static Files
description: Serve directories, single files, and custom filesystem-backed assets.
sidebar_position: 6
---

Zinc supports static assets without leaving `net/http`.

## At a glance

```go
app.Static("/assets", "./public")
app.File("/robots.txt", "./public/robots.txt")
app.Mount("/files", http.FileServer(http.Dir("./public")))
```

Use Zinc's static helpers for normal assets and `Mount` when you already have a standard library handler.

## Serve a directory

```go
if err := app.Static("/assets", "./public"); err != nil {
	log.Fatal(err)
}
```

This mounts a static handler under `/assets`.

## Static options

Use static options to control directory behavior.

```go
if err := app.Static("/docs", "./docs-public",
	zinc.WithStaticIndex("home.html"),
	zinc.WithStaticBrowse(true),
); err != nil {
	log.Fatal(err)
}
```

Available options:

- `WithStaticIndex(index string)`
- `WithStaticBrowse(enabled bool)`

Behavior:

- custom index files are served for directory requests when available
- directory listing works when browse mode is enabled
- unsafe traversal paths are rejected

## Serve a single file

```go
if err := app.File("/robots.txt", "./public/robots.txt"); err != nil {
	log.Fatal(err)
}
```

## Use an `fs.FS`

Zinc works with embedded filesystems and other `fs.FS` implementations.

```go
if err := app.StaticFS("/assets", embeddedAssets); err != nil {
	log.Fatal(err)
}
if err := app.FileFS("/openapi.json", "openapi.json", embeddedAssets); err != nil {
	log.Fatal(err)
}
```

## Mounting existing handlers

If you already have a stdlib handler or third-party file server, you can mount it directly:

```go
app.Mount("/files", http.FileServer(http.Dir("./public")))
```

For most Zinc applications, prefer `Static` and `StaticFS` first. They keep the setup explicit and consistent with the rest of the framework.

## See also

- [Routing](./routing) for mounted handlers.
- [Static middleware](../middleware/static) for middleware-shaped static serving.
- [App API](../api/app) for `Static`, `StaticFS`, `File`, `FileFS`, and `Mount`.
