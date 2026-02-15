#Config: {
	root_path:  string | *"~/zettelkasten"
	index_path: string | *".zk_index"
	graph_path: string | *".zk_graphs"
	todos_path: string | *".zk_todos" // Generated todo lists
	editor:     string | *"nvim"
	folders: {
		fleeting:  string | *"fleeting"
		permanent: string | *"permanent"
		tmp:       string | *"tmp"
	}
}
