import React from "react";
import clsx from "clsx";
import {
  filterDocCardListItems,
  useCurrentSidebarSiblings,
} from "@docusaurus/plugin-content-docs/client";
import DocCard from "@theme/DocCard";

import styles from "./styles.module.css";

function DocCardListForCurrentSidebarCategory({ className }) {
  const items = useCurrentSidebarSiblings();

  return <DocCardList items={items} className={className} />;
}

function DocCardListItem({ item, index }) {
  return (
    <article className={clsx(styles.docCardListItem, "col col--6")}>
      <DocCard item={item} index={index} />
    </article>
  );
}

export default function DocCardList(props) {
  const { items, className } = props;

  if (!items) {
    return <DocCardListForCurrentSidebarCategory {...props} />;
  }

  const filteredItems = filterDocCardListItems(items);

  return (
    <section className={clsx("row", className)}>
      {filteredItems.map((item, index) => (
        <DocCardListItem key={item.href ?? index} item={item} index={index} />
      ))}
    </section>
  );
}
