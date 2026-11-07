package notes

// NoteNode represents a node in the note file tree.
type NoteNode struct {
	Path     string      `json:"path"`     // path relative to workspace root
	IsDir    bool        `json:"is_dir"`   // whether the node is a directory
	Children []*NoteNode `json:"children"` // child nodes (only for directories)
}

// CreateFolderReq is the create folder request.
type CreateFolderReq struct {
	Root string `json:"root"` // absolute workspace root path
	Path string `json:"path"` // new folder path relative to root
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

// SaveNoteImageReq is the save note image request.
type SaveNoteImageReq struct {
	Root     string `json:"root"`      // absolute workspace root path
	NotePath string `json:"note_path"` // note path relative to root
	Data     string `json:"data"`      // base64-encoded original image bytes
}

// SaveNoteImageResp is the save note image response.
type SaveNoteImageResp struct {
	Path string `json:"path"` // image path relative to the note directory
}

// ReadNoteReq is the read note request.
type ReadNoteReq struct {
	Root string `json:"root"` // absolute workspace root path
	Path string `json:"path"` // note path relative to root
}

// ReadNoteImageReq is the read note image request.
type ReadNoteImageReq struct {
	Root     string `json:"root"`      // absolute workspace root path
	NotePath string `json:"note_path"` // note path relative to root
	Path     string `json:"path"`      // image path relative to the note directory
}

// ReadNoteImageResp is the read note image response.
type ReadNoteImageResp struct {
	MIME string `json:"mime"` // detected image MIME type
	Data string `json:"data"` // base64-encoded original image bytes
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
