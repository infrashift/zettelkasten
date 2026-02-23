// Zettel schema with category-specific and type-specific validation
//
// Categories:
//   - untethered: quick captures, project optional
//   - tethered: refined knowledge, project required
//
// Types:
//   - note (default): standard zettel
//   - todo: actionable task with status tracking
//   - daily-note: daily capture note
//   - issue: issue tracking (bug, enhancement, question)
//
// Todo fields (when type=="todo"):
//   - status: required (open, in_progress, closed)
//   - due: optional target date
//   - completed: set automatically when closed
//   - priority: optional (high, medium, low)

#Zettel: #UntetheredZettel | #TetheredZettel

#UntetheredZettel: {
	id:       =~"^[0-9]{14}-[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$" // Enforce YYYYMMDDHHmmss-UUIDv4
	title:    string & !=""
	type:     *"note" | "todo" | "daily-note" | "issue" // Default to "note"
	category: "untethered"
	project?: string // Optional for untethered notes
	tags: [...(string & !="")]
	created: =~"^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}.*" // ISO 8601 required
	parent?: =~"^[0-9]{14}-[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$"                                              // Optional link to parent zettel

	// Todo-specific fields (optional in schema; status required when type=="todo" enforced by code)
	status?:    "open" | "in_progress" | "closed"
	due?:       =~"^[0-9]{4}-[0-9]{2}-[0-9]{2}$" // YYYY-MM-DD
	completed?: =~"^[0-9]{4}-[0-9]{2}-[0-9]{2}$" // YYYY-MM-DD (set when closed)
	priority?:  "high" | "medium" | "low"
}

#TetheredZettel: {
	id:       =~"^[0-9]{14}-[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$" // Enforce YYYYMMDDHHmmss-UUIDv4
	title:    string & !=""
	type:     *"note" | "todo" | "daily-note" | "issue" // Default to "note"
	category: "tethered"
	project:  string & !="" // Required for tethered notes
	tags: [...(string & !="")]
	created: =~"^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}.*" // ISO 8601 required
	parent?: =~"^[0-9]{14}-[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$"                                              // Optional link to parent zettel

	// Todo-specific fields (optional in schema; status required when type=="todo" enforced by code)
	status?:    "open" | "in_progress" | "closed"
	due?:       =~"^[0-9]{4}-[0-9]{2}-[0-9]{2}$" // YYYY-MM-DD
	completed?: =~"^[0-9]{4}-[0-9]{2}-[0-9]{2}$" // YYYY-MM-DD (set when closed)
	priority?:  "high" | "medium" | "low"
}
