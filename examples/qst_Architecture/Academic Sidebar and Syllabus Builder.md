---
scope: primitive
prerequisites:
  - Inverted Graph Index and Degree Metrics
sub_guides: []
tags: [ssg, navigation, syllabus, templates]
---

# Academic Sidebar and Syllabus Builder

In `internal/ssg/navigation.go`, `qst` transforms raw graph relationships and manifest tours into three navigational structures rendered across `layout.html`, `index.html`, and `guide.html`.

## Top-Down Syllabus Generation (`buildSyllabusScopes`)

While the `manifest.yaml` `scopes` list is ordered **bottom-up** (from index `0` `primitive` up to `architecture`), an academic course syllabus is read **top-down**—starting from the overarching system synthesis and drilling down into components and primitives. `buildSyllabusScopes` iterates `config.Scopes` in reverse (`for i := len(config.Scopes) - 1; i >= 0; i--`):

```go
for i := len(config.Scopes) - 1; i >= 0; i-- {
	sc := config.Scopes[i]
	if len(guidesByScope[sc]) > 0 && !seenScopes[sc] {
		orderedScopes = append(orderedScopes, sc)
		seenScopes[sc] = true
	}
}
```

Each syllabus entry displays the guide link, its horizontal prerequisites, and its taxonomy tags.

## Collapsible Sidebar with Session Persistence (`buildSidebar`)

`buildSidebar` constructs two collapsible `<details>` sections in `layout.html`:
1. **Tours (`SidebarTour`)**: Preserves the exact pedagogical sequence defined in `manifest.yaml`.
2. **Scopes (`SidebarScopeGroup`)**: Groups guides alphabetically under each scope tier.

To prevent collapsible `<details data-group-id="...">` groups from snapping shut when navigating between pages, an inline script in `layout.html` serializes the set of open `data-group-id` keys into `sessionStorage` (`qst_sidebar_open_groups`) on every `toggle` event and restores them synchronously during page load.

## Contextual Tour Step Progression (`buildTourSteps`)

For any guide that participates in one or more tours, `buildTourSteps` resolves its 1-based step index (`Step` of `Total`) and computes relative `Prev` and `Next` `GuideRef` pointers, enabling sequential "Previous / Next in Tour" footer cards on every guide page.
