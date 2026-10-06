package notes

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// NotesService provides note operations within a workspace folder.
type NotesService struct{}

func NewNotesService() *NotesService {
	return &NotesService{}
}

// SelectWorkspace selects a system folder as the workspace.
func (s *NotesService) SelectWorkspace(ctx context.Context) (string, error) {
	app := application.Get()
	if app == nil {
		return "", fmt.Errorf("[notes] application not ready")
	}
	selection, err := app.Dialog.OpenFile().
		CanChooseDirectories(true).
		CanChooseFiles(false).
		SetTitle("选择笔记工作区").
		PromptForSingleSelection()
	if err != nil {
		return "", fmt.Errorf("[notes] open directory dialog error: %w", err)
	}
	return selection, nil
}

// ListTree lists the note file tree under a workspace folder.
func (s *NotesService) ListTree(ctx context.Context, root string) ([]*NoteNode, error) {
	root = filepath.Clean(root)
	info, err := os.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("[notes] stat workspace root %s error: %w", root, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("[notes] workspace root %s is not a directory", root)
	}
	return buildTree(root, "")
}

// CreateNote creates a new markdown note.
func (s *NotesService) CreateNote(ctx context.Context, req *CreateNoteReq) error {
	abs, err := resolveSafe(req.Root, req.Path)
	if err != nil {
		return err
	}
	if !strings.HasSuffix(strings.ToLower(abs), ".md") {
		return fmt.Errorf("[notes] note path must end with .md: %s", req.Path)
	}
	if _, err := os.Stat(abs); err == nil {
		return fmt.Errorf("[notes] note already exists: %s", req.Path)
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return fmt.Errorf("[notes] create dir for %s error: %w", req.Path, err)
	}
	name := strings.TrimSuffix(filepath.Base(abs), filepath.Ext(abs))
	content := fmt.Sprintf("# %s\n", name)
	if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
		return fmt.Errorf("[notes] create note %s error: %w", req.Path, err)
	}
	return nil
}

// SaveNote saves the markdown content of a note.
func (s *NotesService) SaveNote(ctx context.Context, req *SaveNoteReq) error {
	abs, err := resolveSafe(req.Root, req.Path)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return fmt.Errorf("[notes] create dir for %s error: %w", req.Path, err)
	}
	if err := os.WriteFile(abs, []byte(req.Content), 0o644); err != nil {
		return fmt.Errorf("[notes] save note %s error: %w", req.Path, err)
	}
	return nil
}

// ReadNote reads the markdown content of a note.
func (s *NotesService) ReadNote(ctx context.Context, req *ReadNoteReq) (string, error) {
	abs, err := resolveSafe(req.Root, req.Path)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		return "", fmt.Errorf("[notes] read note %s error: %w", req.Path, err)
	}
	return string(data), nil
}

// RenameNote renames a note or directory.
func (s *NotesService) RenameNote(ctx context.Context, req *RenameNoteReq) error {
	oldAbs, err := resolveSafe(req.Root, req.OldPath)
	if err != nil {
		return err
	}
	newAbs, err := resolveSafe(req.Root, req.NewPath)
	if err != nil {
		return err
	}
	if _, err := os.Stat(oldAbs); err != nil {
		return fmt.Errorf("[notes] source path not found: %s", req.OldPath)
	}
	if err := os.MkdirAll(filepath.Dir(newAbs), 0o755); err != nil {
		return fmt.Errorf("[notes] create dir for %s error: %w", req.NewPath, err)
	}
	if err := os.Rename(oldAbs, newAbs); err != nil {
		return fmt.Errorf("[notes] rename %s -> %s error: %w", req.OldPath, req.NewPath, err)
	}
	return nil
}

// DeleteNote deletes a note or directory.
func (s *NotesService) DeleteNote(ctx context.Context, req *DeleteNoteReq) error {
	abs, err := resolveSafe(req.Root, req.Path)
	if err != nil {
		return err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return fmt.Errorf("[notes] target path not found: %s", req.Path)
	}
	if info.IsDir() {
		if err := os.RemoveAll(abs); err != nil {
			return fmt.Errorf("[notes] remove dir %s error: %w", req.Path, err)
		}
		return nil
	}
	if err := os.Remove(abs); err != nil {
		return fmt.Errorf("[notes] remove note %s error: %w", req.Path, err)
	}
	return nil
}

// buildTree builds the note file tree recursively.
func buildTree(root, rel string) ([]*NoteNode, error) {
	abs := filepath.Join(root, rel)
	entries, err := os.ReadDir(abs)
	if err != nil {
		return nil, fmt.Errorf("[notes] read dir %s error: %w", rel, err)
	}
	var dirs, files []*NoteNode
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		childRel := filepath.Join(rel, name)
		if entry.IsDir() {
			children, err := buildTree(root, childRel)
			if err != nil {
				return nil, err
			}
			dirs = append(dirs, &NoteNode{
				Path:     filepath.ToSlash(childRel),
				IsDir:    true,
				Children: children,
			})
			continue
		}
		// Expose only markdown files in the tree.
		if !strings.HasSuffix(strings.ToLower(name), ".md") {
			continue
		}
		files = append(files, &NoteNode{
			Path:  filepath.ToSlash(childRel),
			IsDir: false,
		})
	}

	sort.SliceStable(dirs, func(i, j int) bool { return filepath.Base(dirs[i].Path) < filepath.Base(dirs[j].Path) })
	sort.SliceStable(files, func(i, j int) bool { return filepath.Base(files[i].Path) < filepath.Base(files[j].Path) })
	return append(dirs, files...), nil
}

// resolveSafe resolves a relative path to an absolute path.
func resolveSafe(root, rel string) (string, error) {
	root = filepath.Clean(root)
	cleaned := filepath.Clean(filepath.FromSlash(rel))
	if filepath.IsAbs(cleaned) || strings.HasPrefix(cleaned, "..") {
		return "", fmt.Errorf("[notes] path escapes workspace root: %s", rel)
	}
	abs := filepath.Join(root, cleaned)
	rel2, err := filepath.Rel(root, abs)
	if err != nil || strings.HasPrefix(filepath.ToSlash(rel2), "..") {
		return "", fmt.Errorf("[notes] path escapes workspace root: %s", rel)
	}
	return abs, nil
}
