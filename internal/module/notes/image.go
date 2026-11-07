package notes

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// maxNoteImageSize is the maximum size of a note image in bytes.
const maxNoteImageSize = 10 << 20

// SaveNoteImage saves a note image.
func (s *NotesService) SaveNoteImage(ctx context.Context, req *SaveNoteImageReq) (*SaveNoteImageResp, error) {
	if req == nil {
		return nil, fmt.Errorf("[notes] save note image request is nil")
	}
	noteAbs, err := resolveImageNote(req.Root, req.NotePath)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Data) == "" {
		return nil, fmt.Errorf("[notes] note image data is empty")
	}
	// reject oversize base64 before decoding to avoid allocating a huge buffer.
	if len(req.Data) > base64.StdEncoding.EncodedLen(maxNoteImageSize) {
		return nil, fmt.Errorf("[notes] note image exceeds %d bytes", maxNoteImageSize)
	}
	data, err := base64.StdEncoding.DecodeString(req.Data)
	if err != nil {
		return nil, fmt.Errorf("[notes] decode note image data error: %w", err)
	}
	if len(data) > maxNoteImageSize {
		return nil, fmt.Errorf("[notes] note image exceeds %d bytes", maxNoteImageSize)
	}
	_, ext, err := detectNoteImageFormat(data)
	if err != nil {
		return nil, err
	}

	assetsAbs, err := resolveSafe(req.Root, ".assets")
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(assetsAbs, 0o755); err != nil {
		return nil, fmt.Errorf("[notes] write note image assets directory error: %w", err)
	}
	assetsReal, err := resolveExistingPathWithinRoot(req.Root, assetsAbs)
	if err != nil {
		return nil, err
	}
	hash := fmt.Sprintf("%x", sha256.Sum256(data))
	name := hash + ext
	imageAbs := filepath.Join(assetsReal, name)
	if info, statErr := os.Lstat(imageAbs); statErr == nil && info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("[notes] note image path must not be a symbolic link: %s", name)
	}
	if err := writeNoteImage(imageAbs, data); err != nil {
		return nil, err
	}
	logicalImageAbs := filepath.Join(assetsAbs, name)
	rel, err := filepath.Rel(filepath.Dir(noteAbs), logicalImageAbs)
	if err != nil {
		return nil, fmt.Errorf("[notes] resolve note image relative path error: %w", err)
	}
	return &SaveNoteImageResp{Path: filepath.ToSlash(rel)}, nil
}

// ReadNoteImage reads a note image.
func (s *NotesService) ReadNoteImage(ctx context.Context, req *ReadNoteImageReq) (*ReadNoteImageResp, error) {
	if req == nil {
		return nil, fmt.Errorf("[notes] read note image request is nil")
	}
	if _, err := resolveImageNote(req.Root, req.NotePath); err != nil {
		return nil, err
	}
	if err := validateLocalRelativePath(req.Path); err != nil {
		return nil, err
	}

	noteRel := filepath.Clean(filepath.FromSlash(req.NotePath))
	imageRel := filepath.Join(filepath.Dir(noteRel), filepath.FromSlash(req.Path))
	imageAbs, err := resolveSafe(req.Root, imageRel)
	if err != nil {
		return nil, err
	}
	imageAbs, err = resolveExistingPathWithinRoot(req.Root, imageAbs)
	if err != nil {
		return nil, err
	}
	data, err := readNoteImage(imageAbs)
	if err != nil {
		return nil, fmt.Errorf("[notes] read note image %s error: %w", req.Path, err)
	}
	mime, _, err := detectNoteImageFormat(data)
	if err != nil {
		return nil, err
	}
	return &ReadNoteImageResp{
		MIME: mime,
		Data: base64.StdEncoding.EncodeToString(data),
	}, nil
}

// resolveImageNote resolves the image note path.
func resolveImageNote(root, notePath string) (string, error) {
	if strings.TrimSpace(root) == "" {
		return "", fmt.Errorf("[notes] workspace root path is empty")
	}
	info, err := os.Stat(root)
	if err != nil {
		return "", fmt.Errorf("[notes] stat workspace root path %s error: %w", root, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("[notes] workspace root path %s is not a directory", root)
	}
	if strings.TrimSpace(notePath) == "" {
		return "", fmt.Errorf("[notes] note path is empty")
	}
	abs, err := resolveSafe(root, notePath)
	if err != nil {
		return "", err
	}
	if !strings.EqualFold(filepath.Ext(abs), ".md") {
		return "", fmt.Errorf("[notes] note path must end with .md: %s", notePath)
	}
	return abs, nil
}

// validateLocalRelativePath validates a path is local and relative.
func validateLocalRelativePath(path string) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("[notes] path is empty")
	}
	if strings.ContainsAny(path, "?#") {
		return fmt.Errorf("[notes] path must not contain a query or fragment: %s", path)
	}
	parsed, err := url.Parse(path)
	if err != nil {
		return fmt.Errorf("[notes] invalid path %s: %w", path, err)
	}
	localPath := filepath.FromSlash(path)
	if parsed.Scheme != "" || parsed.Host != "" || filepath.IsAbs(localPath) || filepath.VolumeName(localPath) != "" || strings.HasPrefix(path, "/") || strings.HasPrefix(path, `\`) {
		return fmt.Errorf("[notes] path must be local and relative: %s", path)
	}
	return nil
}

// detectNoteImageFormat detects the note image format.
func detectNoteImageFormat(data []byte) (mime, ext string, err error) {
	if len(data) == 0 {
		return "", "", fmt.Errorf("[notes] note image content is empty")
	}
	mime = http.DetectContentType(data)
	switch mime {
	case "image/png":
		return mime, ".png", nil
	case "image/jpeg":
		return mime, ".jpg", nil
	case "image/gif":
		return mime, ".gif", nil
	case "image/webp":
		return mime, ".webp", nil
	default:
		return "", "", fmt.Errorf("[notes] unsupported note image format: %s", mime)
	}
}

// resolveExistingPathWithinRoot resolves the existing path within the root.
func resolveExistingPathWithinRoot(root, target string) (string, error) {
	resolvedRoot, err := resolveRealPath(root)
	if err != nil {
		return "", fmt.Errorf("[notes] resolve workspace root path %s error: %w", root, err)
	}
	resolvedTarget, err := resolveRealPath(target)
	if err != nil {
		return "", fmt.Errorf("[notes] resolve target path %s error: %w", target, err)
	}
	if !isPathWithinRoot(resolvedRoot, resolvedTarget) {
		return "", fmt.Errorf("[notes] path escapes workspace root path: %s", target)
	}
	return resolvedTarget, nil
}

// isPathWithinRoot reports whether target path stays inside root path.
func isPathWithinRoot(root, target string) bool {
	rel, err := filepath.Rel(filepath.Clean(root), filepath.Clean(target))
	if err != nil || filepath.IsAbs(rel) || rel == ".." {
		return false
	}
	return !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// writeNoteImage writes a note image without overwriting different content.
func writeNoteImage(path string, data []byte) error {
	file, err := os.CreateTemp(filepath.Dir(path), ".image-*")
	if err != nil {
		return fmt.Errorf("[notes] create temporary note image error: %w", err)
	}
	tmpPath := file.Name()
	complete := false
	defer func() {
		_ = file.Close()
		if !complete {
			_ = os.Remove(tmpPath)
		}
	}()
	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("[notes] write note image %s error: %w", path, err)
	}
	if err := file.Sync(); err != nil {
		return fmt.Errorf("[notes] sync note image %s error: %w", path, err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("[notes] close note image %s error: %w", path, err)
	}
	if err := os.Chmod(tmpPath, 0o644); err != nil {
		return fmt.Errorf("[notes] chmod note image %s error: %w", path, err)
	}

	if existing, readErr := readNoteImage(path); readErr == nil && bytes.Equal(existing, data) {
		return nil
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("[notes] replace note image %s error: %w", path, err)
	}
	complete = true
	return nil
}

// readNoteImage reads a note image within the maximum size limit.
func readNoteImage(path string) ([]byte, error) {
	image, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer image.Close()

	data, err := io.ReadAll(io.LimitReader(image, maxNoteImageSize+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxNoteImageSize {
		return nil, fmt.Errorf("note image exceeds %d bytes", maxNoteImageSize)
	}
	return data, nil
}
