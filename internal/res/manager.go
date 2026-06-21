package res

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Manager struct {
	Root       string
	ClientInfo ClientInfo
	FoundFiles []string
	Archives   []*GRF
}

func NewManager(root string) (*Manager, error) {
	if root == "" {
		return nil, errors.New("empty root")
	}

	m := &Manager{Root: filepath.Clean(root)}
	m.scanKnownFiles()
	m.ClientInfo = ClientInfo{
		Connections: []Connection{
			{Display: "Local rAthena", Address: "127.0.0.1", Port: 6900, Version: 55, LangType: 0},
		},
	}

	if source, data, ok := m.ReadFirst(clientInfoCandidates); ok {
		info, err := ParseClientInfo(data)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", source, err)
		}
		if len(info.Connections) > 0 {
			m.ClientInfo = info
		}
	}

	return m, nil
}

func (m *Manager) Find(name string) (string, bool) {
	normalized := normalizePath(name)
	candidates := []string{
		filepath.Join(m.Root, normalized),
		filepath.Join(m.Root, strings.ReplaceAll(normalized, "\\", string(filepath.Separator))),
		filepath.Join(m.Root, strings.ReplaceAll(normalized, "/", string(filepath.Separator))),
	}

	for _, candidate := range candidates {
		if stat, err := os.Stat(candidate); err == nil && !stat.IsDir() {
			return candidate, true
		}
	}
	return "", false
}

func (m *Manager) ReadFile(name string) ([]byte, error) {
	path, ok := m.Find(name)
	if ok {
		return os.ReadFile(path)
	}

	for _, archive := range m.Archives {
		data, err := archive.ReadFile(name)
		if err == nil {
			return data, nil
		}
		if errors.Is(err, ErrGRFNotFound) {
			for _, match := range archive.NamesWithSuffix(name) {
				data, err := archive.ReadFile(match)
				if err == nil {
					return data, nil
				}
				if !errors.Is(err, ErrGRFNotFound) {
					return nil, err
				}
			}
			continue
		}
		return nil, err
	}
	return nil, fmt.Errorf("resource not found: %s", name)
}

func (m *Manager) FindFirst(names []string) (string, bool) {
	for _, name := range names {
		if path, ok := m.Find(name); ok {
			return path, true
		}
	}
	return "", false
}

func (m *Manager) ReadFirst(names []string) (string, []byte, bool) {
	for _, name := range names {
		data, err := m.ReadFile(name)
		if err == nil {
			if path, ok := m.Find(name); ok {
				return path, data, true
			}
			return name, data, true
		}
	}
	return "", nil, false
}

func (m *Manager) scanKnownFiles() {
	for _, name := range append(clientInfoCandidates, "data.grf", "rdata.grf", "fdata.grf", "event.grf") {
		if path, ok := m.Find(name); ok {
			m.FoundFiles = append(m.FoundFiles, path)
		}
	}

	archivePaths := make([]string, 0)
	seen := make(map[string]struct{})
	if entries, err := os.ReadDir(m.Root); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			ext := strings.ToLower(filepath.Ext(name))
			if ext != ".grf" && ext != ".gpf" {
				continue
			}
			path := filepath.Join(m.Root, name)
			archivePaths = append(archivePaths, path)
			seen[strings.ToLower(path)] = struct{}{}
		}
	}
	for _, name := range []string{"data.grf", "rdata.grf", "fdata.grf", "event.grf"} {
		path := filepath.Join(m.Root, name)
		if _, ok := seen[strings.ToLower(path)]; ok {
			continue
		}
		if _, err := os.Stat(path); err != nil {
			continue
		}
		archivePaths = append(archivePaths, path)
	}
	sort.SliceStable(archivePaths, func(i, j int) bool {
		return archivePriority(archivePaths[i]) < archivePriority(archivePaths[j])
	})
	for _, path := range archivePaths {
		archive, err := OpenGRF(path)
		if err != nil {
			continue
		}
		m.Archives = append(m.Archives, archive)
	}
}

func archivePriority(path string) string {
	name := strings.ToLower(filepath.Base(path))
	switch name {
	case "data.grf":
		return "z-data.grf"
	case "rdata.grf":
		return "y-rdata.grf"
	case "fdata.grf":
		return "x-fdata.grf"
	default:
		return name
	}
}

func normalizePath(name string) string {
	name = strings.TrimPrefix(name, "./")
	name = strings.TrimPrefix(name, ".\\")
	return filepath.Clean(name)
}

var clientInfoCandidates = []string{
	"data/clientinfo.xml",
	"data/sclientinfo.xml",
	"clientinfo.xml",
	"sclientinfo.xml",
	"System/clientinfo.xml",
	"System/sclientinfo.xml",
}
