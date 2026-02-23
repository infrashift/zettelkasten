// Template schema definitions
// All templates must produce zettels that conform to #Zettel

// Template metadata
#TemplateMeta: {
	name:        string & !=""
	description: string
	category:    "untethered" | "tethered"
	tags:        [...(string & !="")]
}

// Standard templates
#MeetingTemplate: #TemplateMeta & {
	name:        "meeting"
	description: "Meeting notes with attendees and action items"
	category:    "untethered"
	tags:        ["meeting"]
}

#BookReviewTemplate: #TemplateMeta & {
	name:        "book-review"
	description: "Book review with rating and key takeaways"
	category:    "tethered"
	tags:        ["book", "review"]
}

#SnippetTemplate: #TemplateMeta & {
	name:        "snippet"
	description: "Code snippet with context and explanation"
	category:    "untethered"
	tags:        ["code", "snippet"]
}

// Custom templates
#ProjectIdeaTemplate: #TemplateMeta & {
	name:        "project-idea"
	description: "Project idea with goals and next steps"
	category:    "untethered"
	tags:        ["idea", "project"]
}

#UserStoryTemplate: #TemplateMeta & {
	name:        "user-story"
	description: "User story in standard format with acceptance criteria"
	category:    "untethered"
	tags:        ["user-story", "requirements"]
}

#FeatureTemplate: #TemplateMeta & {
	name:        "feature"
	description: "Feature specification with requirements and design notes"
	category:    "untethered"
	tags:        ["feature", "spec"]
}
