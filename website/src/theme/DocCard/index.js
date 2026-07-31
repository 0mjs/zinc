import React from "react";
import clsx from "clsx";
import Link from "@docusaurus/Link";
import {
  findFirstSidebarItemLink,
  useDocById,
} from "@docusaurus/plugin-content-docs/client";
import { usePluralForm } from "@docusaurus/theme-common";
import { translate } from "@docusaurus/Translate";
import Heading from "@theme/Heading";

import styles from "./styles.module.css";

function useCategoryItemsPlural() {
  const { selectMessage } = usePluralForm();

  return (count) =>
    selectMessage(
      count,
      translate(
        {
          message: "1 item|{count} items",
          id: "theme.docs.DocCard.categoryDescription.plurals",
          description:
            "The default description for a category card in a generated index",
        },
        { count },
      ),
    );
}

function CardLayout({ className, href, index, title, description }) {
  const marker = String(index + 1).padStart(2, "0");

  return (
    <Link
      href={href}
      className={clsx("card padding--lg", styles.cardContainer, className)}
    >
      <Heading as="h2" className={styles.cardTitle} title={title}>
        <span className={styles.cardMarker} aria-hidden="true">
          {marker}
        </span>
        <span>{title}</span>
      </Heading>
      {description && (
        <p className={styles.cardDescription} title={description}>
          {description}
        </p>
      )}
    </Link>
  );
}

function CardCategory({ item, index }) {
  const href = findFirstSidebarItemLink(item);
  const categoryItemsPlural = useCategoryItemsPlural();

  if (!href) {
    return null;
  }

  return (
    <CardLayout
      className={item.className}
      href={href}
      index={index}
      title={item.label}
      description={item.description ?? categoryItemsPlural(item.items.length)}
    />
  );
}

function CardLink({ item, index }) {
  const doc = useDocById(item.docId ?? undefined);

  return (
    <CardLayout
      className={item.className}
      href={item.href}
      index={index}
      title={item.label}
      description={item.description ?? doc?.description}
    />
  );
}

export default function DocCard({ item, index = 0 }) {
  switch (item.type) {
    case "link":
      return <CardLink item={item} index={index} />;
    case "category":
      return <CardCategory item={item} index={index} />;
    default:
      throw new Error(`unknown item type ${JSON.stringify(item)}`);
  }
}
