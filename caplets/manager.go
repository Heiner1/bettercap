package caplets

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/evilsocket/islazy/fs"
	"github.com/evilsocket/islazy/str"
)

var (
	cache     = make(map[string]*Caplet)
	cacheLock = sync.Mutex{}
)

func List() []Caplet {
	caplets := make([]Caplet, 0)
	for _, searchPath := range LoadPaths {
		files, _ := filepath.Glob(searchPath + "/*" + Suffix)
		files2, _ := filepath.Glob(searchPath + "/*/*" + Suffix)

		for _, fileName := range append(files, files2...) {
			if stats, err := os.Stat(fileName); err == nil {
				base := strings.Replace(fileName, searchPath+"/", "", -1)
				base = strings.Replace(base, Suffix, "", -1)

				caplets = append(caplets, Caplet{
					Name: base,
					Path: fileName,
					Size: stats.Size(),
				})
			}
		}
	}

	sort.Slice(caplets, func(i, j int) bool {
		return strings.Compare(caplets[i].Name, caplets[j].Name) == -1
	})

	return caplets
}

func Load(name string) (error, *Caplet) {
	cacheLock.Lock()
	defer cacheLock.Unlock()

	if caplet, found := cache[name]; found {
		return nil, caplet
	}

	names := []string{}
	if !strings.HasSuffix(name, Suffix) {
		name += Suffix
	}

	for _, path := range LoadPaths {
		names = append(names, filepath.Join(path, name))
	}

	for _, filename := range names {
		if fs.Exists(filename) {
			cap := &Caplet{
				Path: filename,
				Code: make([]string, 0),
			}

			input, err := os.Open(filename)
			if err != nil {
				return fmt.Errorf("error reading caplet %s: %v", filename, err), nil
			}
			defer input.Close()

			scanner := bufio.NewScanner(input)
			scanner.Split(bufio.ScanLines)
			for scanner.Scan() {
				line := str.Trim(scanner.Text())
				if line == "" || line[0] == '#' {
					continue
				}
				cap.Code = append(cap.Code, line)
			}

			cache[name] = cap
			return nil, cap
		}
	}

	return fmt.Errorf("caplet %s not found", name), nil
}
