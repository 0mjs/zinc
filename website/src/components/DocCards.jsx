import React from "react";
import Link from "@docusaurus/Link";

export default function DocCards({ cards }) {
  return (
    <div className="zinc-doc-cards">
      {cards.map((card) => (
        <Link key={card.to} to={card.to} className="zinc-doc-card">
          <strong>{card.title}</strong>
          <span>{card.description}</span>
        </Link>
      ))}
    </div>
  );
}
