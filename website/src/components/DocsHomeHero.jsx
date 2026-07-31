import React from "react";
import Link from "@docusaurus/Link";
import { Highlight } from "prism-react-renderer";

const { nordDark } = require("../prism/nord");

const serverExample = `package main

import (
    "log"

    "github.com/0mjs/zinc"
)

func main() {
    app := zinc.New()

    app.Get("/", func(c *zinc.Context) error {
        return c.String("Hello from Zinc!")
    })

    log.Fatal(app.Listen(":8080"))
}`;

export default function DocsHomeHero() {
  return (
    <section className="zinc-home-hero">
      <div className="zinc-home-hero__copy">
        <div className="zinc-kicker">
          <span className="zinc-kicker__mark" aria-hidden="true" />
          Zinc documentation
        </div>

        <h1>
          Fast Go APIs.<br />
          Still <span className="zinc-home-hero__protocol">net/http</span>.
        </h1>

        <p className="zinc-home-hero__lead">
          Expressive routing and practical helpers, built on the HTTP stack Go
          developers already know.
        </p>

        <div className="zinc-home-hero__actions">
          <Link className="zinc-button zinc-button--primary" to="/getting-started/quick-start">
            Build your first API
            <span aria-hidden="true">→</span>
          </Link>
          <Link className="zinc-button zinc-button--secondary" to="/guide/routing">
            Read the guide
          </Link>
        </div>

        <ul className="zinc-home-hero__traits" aria-label="Zinc highlights">
          <li><span aria-hidden="true">✓</span> Standard library compatible</li>
          <li><span aria-hidden="true">✓</span> Small, explicit API</li>
          <li><span aria-hidden="true">✓</span> Built for speed</li>
        </ul>
      </div>

      <div className="zinc-code-window" aria-label="A minimal Zinc server">
        <div className="zinc-code-window__bar">
          <span className="zinc-code-window__dots" aria-hidden="true">
            <i /><i /><i />
          </span>
          <span>main.go</span>
          <span className="zinc-code-window__badge">
            <img src="/img/go.png" alt="Go" />
          </span>
        </div>
        <Highlight theme={nordDark} code={serverExample} language="go">
          {({ className, style, tokens, getLineProps, getTokenProps }) => (
            <pre
              className={`${className} zinc-code-window__source`}
              style={{ ...style, backgroundColor: "transparent" }}
            >
              <code>
                {tokens.map((line, lineIndex) => (
                  <span
                    key={lineIndex}
                    {...getLineProps({ line })}
                    className="zinc-code-window__line"
                  >
                    {line.map((token, tokenIndex) => (
                      <span key={tokenIndex} {...getTokenProps({ token })} />
                    ))}
                    {"\n"}
                  </span>
                ))}
              </code>
            </pre>
          )}
        </Highlight>
        <div className="zinc-code-window__footer">
          <span>$</span>
          <code>go get github.com/0mjs/zinc</code>
        </div>
      </div>
    </section>
  );
}
