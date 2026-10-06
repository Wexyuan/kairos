package notes

// NoteNode represents a node in the note file tree.
type NoteNode struct {
	Path     string      `json:"path"`     // path relative to workspace root
	IsDir    bool        `json:"is_dir"`   // whether the node is a directory
	Children []*NoteNode `json:"children"` // child nodes (only for directories)
}

// CreateNoteReq is the create note request.
type CreateNoteReq struct {
	Root string `json:"root"` // absolute workspace root path
	Path string `json:"path"` // new note path relative to root
}

// SaveNoteReq is the save note request.
type SaveNoteReq struct {
	Root    string `json:"root"`    // absolute workspace root path
	Path    string `json:"path"`    // note path relative to root
	Content string `json:"content"` // markdown content
}

// ReadNoteReq is the read note request.
type ReadNoteReq struct {
	Root string `json:"root"` // absolute workspace root path
	Path string `json:"path"` // note path relative to root
}

// RenameNoteReq is the rename note request.
type RenameNoteReq struct {
	Root    string `json:"root"`     // absolute workspace root path
	OldPath string `json:"old_path"` // old path relative to root
	NewPath string `json:"new_path"` // new path relative to root
}

// DeleteNoteReq is the delete note request.
type DeleteNoteReq struct {
	Root string `json:"root"` // absolute workspace root path
	Path string `json:"path"` // note path relative to root
}
