// Zettel schema with category-specific and type-specific validation
//
// Categories:
//   - fleeting: quick captures, project optional
//   - permanent: refined knowledge, project required
//
// Types:
//   - note (default): standard zettel
//   - todo: actionable task with status tracking
//   - dailynote: daily capture note
//
// Links:
//   - Any zettel can link to any other zettel via the links field
//   - Links are zettel IDs that this note references
//
// Todo fields (when type=="todo"):
//   - status: required (open, in_progress, closed)
//   - due: optional target date
//   - completed: set automatically when closed
//   - priority: optional (high, medium, low)

#Zettel: #FleetingZettel | #PermanentZettel

#FleetingZettel: {
	id:       =~"^[0-9]{12}$" // Enforce YYYYMMDDHHMM
	title:    string & !=""
	type:     *"note" | "todo" | "dailynote" // Default to "note"
	category: "fleeting"
	project?: string // Optional for fleeting notes
	tags: [...(string & !="")]
	created: =~"^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}.*" // ISO 8601 required
	parent?: =~"^[0-9]{12}$"                                              // Optional link to parent zettel
	links?: [...=~"^[0-9]{12}$"]                                          // Links to other zettels

	// Todo-specific fields (optional in schema; status required when type=="todo" enforced by code)
	status?:    "open" | "in_progress" | "closed"
	due?:       =~"^[0-9]{4}-[0-9]{2}-[0-9]{2}$" // YYYY-MM-DD
	completed?: =~"^[0-9]{4}-[0-9]{2}-[0-9]{2}$" // YYYY-MM-DD (set when closed)
	priority?:  "high" | "medium" | "low"
}

#PermanentZettel: {
	id:       =~"^[0-9]{12}$" // Enforce YYYYMMDDHHMM
	title:    string & !=""
	type:     *"note" | "todo" | "dailynote" // Default to "note"
	category: "permanent"
	project:  string & !="" // Required for permanent notes
	tags: [...(string & !="")]
	created: =~"^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}.*" // ISO 8601 required
	parent?: =~"^[0-9]{12}$"                                              // Optional link to parent zettel
	links?: [...=~"^[0-9]{12}$"]                                          // Links to other zettels

	// Todo-specific fields (optional in schema; status required when type=="todo" enforced by code)
	status?:    "open" | "in_progress" | "closed"
	due?:       =~"^[0-9]{4}-[0-9]{2}-[0-9]{2}$" // YYYY-MM-DD
	completed?: =~"^[0-9]{4}-[0-9]{2}-[0-9]{2}$" // YYYY-MM-DD (set when closed)
	priority?:  "high" | "medium" | "low"
}
