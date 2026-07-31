import React from "react";
import Link from "@docusaurus/Link";

export default function DocCards({ cards }) {
  return (
    <div className="zinc-doc-cards">
      {cards.map((card, index) => (
        <Link key={card.to} to={card.to} className="zinc-doc-card">
          <span className="zinc-doc-card__number" aria-hidden="true">
            {String(index + 1).padStart(2, "0")}
          </span>
          <span className="zinc-doc-card__content">
            <strong>{card.title}</strong>
            <span>{card.description}</span>
          </span>
          <span className="zinc-doc-card__arrow" aria-hidden="true">→</span>
        </Link>
      ))}
    </div>
  );
}
