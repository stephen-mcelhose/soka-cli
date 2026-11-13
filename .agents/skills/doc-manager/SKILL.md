---
name: doc-manager
description: "Manage the soka-cli documentation corpus: README.md, AGENTS.md, docs/{research,adrs}, and docs/wiki/{concepts,definitions,examples}. Use when creating, updating, querying, or linting project documents. Trigger on: 'update the docs', 'add to the wiki', 'write an ADR', 'write a definition', 'write a concept', 'ingest this into the wiki', 'lint the docs', or any explicit mention of the doc structure."
---

# Doc Manager

Manages the soka-cli documentation corpus. Operates at two levels:

1. **Project docs** — `README.md`, `AGENTS.md`, `docs/research/`, `docs/adrs/`
2. **Wiki** — `docs/wiki/{concepts/,definitions/,examples/}` — an interlinked knowledge base about the domain

The LLM is the author; the docs are the codebase.

---

## Directory layout

```
README.md                        # project overview, install, quickstart
AGENTS.md                        # agent conventions for this repo
docs/
  research/                      # research notes, experiment write-ups, findings
  adrs/                          # architectural decision records
  wiki/
    concepts/                    # domain concepts — interlinked, schema below
    definitions/                 # precise term definitions — strict schema
    examples/                    # code and/or math illustrations
    index.md                     # auto-maintained index of all wiki pages
    log.md                       # append-only operation log
```

---

## OKF Frontmatter (required on every wiki doc)

All files under `docs/wiki/` MUST carry OKF frontmatter:

```yaml
---
type: <concept | definition | example>
title: <Title Case>
description: <one sentence — what this page covers and why it matters>
tags: [<relevant>, <keywords>]
timestamp: <ISO-8601 UTC — run: date -u +%Y-%m-%dT%H:%M:%SZ>
---
```

`resource` is optional (omit if the page has no single canonical source).

Research notes and ADRs use a lighter header (see schemas below); they do not use OKF.

---

## Wiki schemas

### definitions/ — MUST follow this schema

A definition page states exactly what a term means in this project's context.
Every definition MUST have all of these sections:

```markdown
---
type: definition
title: <Term>
description: <one-sentence plain-English gloss>
tags: [...]
timestamp: ...
---

# <Term>

**Definition:** <Precise, unambiguous statement of what the term means in this codebase/domain.>

## Formal statement (if applicable)

<Equation, pseudocode, or formal notation. Use LaTeX fencing: $...$ inline, $$...$$ block.>

## Distinctions

<How this term differs from related terms it is commonly confused with. At least one entry.>

## See also

- [[ConceptName]] — <why related>
- [[OtherDefinition]] — <why related>

## Sources

- <citation or path>
```

### concepts/ — MAY follow this schema

A concept page synthesises an idea, mechanism, or design pattern. Schema is recommended but flexible:

```markdown
---
type: concept
title: <Concept Name>
description: <one-sentence summary>
tags: [...]
timestamp: ...
---

# <Concept Name>

<2–4 paragraph synthesis. Lead with the core intuition, then detail.>

## Key properties

- ...

## Relationship to other concepts

Concepts MUST interlink with related concepts and definitions using [[wiki-links]]:
- [[RelatedConcept]] — <brief note on relationship>
- [[TermDefinition]] — <brief note on relationship>

## Examples

<Optional. If substantive, link to an examples/ page instead: [[ExampleName]]>

## Sources

- <citation or path>
```

**[[wiki-links]] rule:** Every concept page MUST use `[[PageTitle]]` syntax to reference
related concepts and definitions. The link text is the exact title of the target page.
Links resolve within `docs/wiki/` — do not use relative paths in concept cross-references.

### examples/ — code and/or math

An example page illustrates a concept or definition concretely. No strict schema, but MUST:
- Carry OKF frontmatter (`type: example`)
- Reference the concept or definition it illustrates with `[[wiki-links]]`
- Contain at least one of: a runnable code block, a worked mathematical derivation, or a diagram

```markdown
---
type: example
title: <Descriptive title>
description: <what this example demonstrates>
tags: [...]
timestamp: ...
---

# <Title>

Illustrates: [[ConceptOrDefinitionName]]

<Code block or math derivation with explanation.>
```

---

## Project doc schemas

### docs/adrs/NNN-title.md

```markdown
# ADR-NNN: <Title>

**Status:** <proposed | accepted | deprecated | superseded by ADR-NNN>
**Date:** YYYY-MM-DD

## Context

<What situation or force drove this decision?>

## Decision

<What was decided, stated precisely.>

## Consequences

<What becomes easier, what becomes harder, what is now a constraint.>
```

### docs/research/slug.md

Freeform. SHOULD include: objective, method, findings, open questions.

---

## Operations

### create

Write a new document at the correct path. Apply the schema for its type.
For wiki pages: append a row to `docs/wiki/index.md` and an entry to `docs/wiki/log.md`.

### update

Read the existing doc first. Edit it in-place. For wiki pages: propagate to pages
that link to or should link to the updated page.

### query

Read `docs/wiki/index.md`. Identify 3–6 relevant pages. Read them. Follow `[[wiki-links]]`
to related pages if relevant. Synthesise an answer, citing each source page.

### lint

Check all `docs/wiki/**/*.md` for:
- Missing or incomplete OKF frontmatter
- Concepts missing `[[wiki-links]]` to clearly related pages
- Definitions missing required sections
- Broken `[[wiki-links]]` (target page does not exist)
- Pages missing from `docs/wiki/index.md`
- Orphaned pages (no inbound links from any other wiki page)

Fix what can be fixed automatically. Report what requires human judgment.

### ingest

Read a source (URL or file). Write a summary page in the appropriate wiki subdirectory.
Propagate to related pages. Update `index.md` and `log.md`.

---

## [[wiki-links]] resolution

`[[PageTitle]]` resolves to the wiki page whose `title` frontmatter field exactly matches
`PageTitle`. To find the file: read `docs/wiki/index.md` or `Glob("**/*.md", "docs/wiki/")`,
then match on the `title` field.

When *writing* `[[wiki-links]]`, use the exact `title` from the target page's frontmatter.
When *reading* a page with `[[wiki-links]]`, resolve them to find the actual file for reading.

---

## Rules

- **Never modify raw source files.** Read them; never write to them.
- **index.md is always current.** Every wiki write updates it before finishing.
- **log.md is append-only.** Never rewrite or delete existing entries.
- **OKF is mandatory on all wiki docs.** Non-compliant pages are a lint error.
- **Definitions are strict.** A definition page missing any required section is incomplete.
- **Concepts interlink.** A concept page with no `[[wiki-links]]` is a lint warning.
- **Examples ground concepts.** An example with no reference to a concept/definition is a lint warning.
- **ADR numbers are sequential.** Check existing ADRs before assigning a number.
