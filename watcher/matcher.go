package watcher

import (
	"path/filepath"
	"regexp"
	"strings"
)

var escRe = regexp.MustCompile(`[\/\.\$\^]`)

func str2regex(str string) *regexp.Regexp {
	//placeholder "matcher" syntax
	glob := strings.Replace(str, `/**/`, `GLOB!ALL`, -1)
	glob = strings.Replace(glob, `*`, `GLOB!STAR`, -1)
	//escape regex syntax
	glob = escRe.ReplaceAllStringFunc(glob, func(char string) string {
		return `\` + char
	})
	//convert "matcher" syntax into regex syntax
	glob = strings.Replace(glob, `GLOB!STAR`, `[^\/]+`, -1)
	glob = strings.Replace(glob, `GLOB!ALL`, `\/([^\/]+\/)*`, -1)
	glob = "^" + glob + "$"
	return regexp.MustCompile(glob)
}

//matcher is configured to match files
type matcher struct {
	hidden    bool
	include   bool
	str       string
	file, dir *regexp.Regexp
}

func (m *matcher) glob(str string) {
	m.str = str
	m.file = str2regex(str)
	i := strings.LastIndex(str, "/")

	m.dir = str2regex(str[:i+1])
}

func (m *matcher) flip(b bool) bool {
	if !m.include {
		return !b
	}
	return b
}

func (m *matcher) matchFile(s string) bool {
	return m.match(true, s)
}

func (m *matcher) matchDir(s string) bool {
	return m.match(false, s)
}

func (m *matcher) match(isFile bool, s string) bool {
	//generally, you dont want to watch the entire .git directory tree
	if !m.hidden && strings.HasPrefix(filepath.Base(s), ".") {
		return false
	}
	//file match
	if m.file != nil && isFile {
		return m.flip(m.file.MatchString(s))
	}
	//directory match (only exclude files)
	if m.include && m.dir != nil && !isFile {
		//directories must have trailing slash
		if !strings.HasSuffix(s, "/") {
			s += "/"
		}
		//include child directories
		if strings.HasPrefix(m.str, s) {
			return true
		}
		//directory match
		return m.flip(m.dir.MatchString(s))
	}
	return true
}
