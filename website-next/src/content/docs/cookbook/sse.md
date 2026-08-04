---
title: Server-Sent Events
description: Stream structured server events over a long-lived HTTP response.
---

```go
app.Get("/events", func(c *zinc.Context) error {
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()

    for {
        select {
        case now := <-ticker.C:
            if err := c.SSE(zinc.SSEvent{
                Event: "clock",
                ID:    strconv.FormatInt(now.Unix(), 10),
                Data:  zinc.Map{"time": now.UTC()},
            }); err != nil {
                return err
            }
        case <-c.Context().Done():
            return nil
        }
    }
})
```

Browser client:

```js
const events = new EventSource("/events");
events.addEventListener("clock", (event) => {
  console.log(JSON.parse(event.data));
});
```

`c.SSE` sets the event-stream headers and flushes each event when the writer supports flushing.
