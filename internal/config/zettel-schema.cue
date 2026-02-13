#Zettel: {
    id:       =~"^[0-9]{12}$"  // Enforce YYYYMMDDHHMM
    title:    string & !=""
    project:  string
    category: "fleeting" | "permanent"
    tags:     [...string]
}