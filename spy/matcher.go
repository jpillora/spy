package spy

import (
	"path/filepath"
	"regexp"
	"strings"
)

var escRe = regexp.MustCompile(`[\/\.\$\^]`)

func str2regex(str string) *regexp.Regexp {
	//"matcher" syntax into placeholders
	glob := strings.Replace(str, `/**/`, `GLOB!ALL`, -1)
	glob = strings.Replace(glob, `*`, `GLOB!STAR`, -1)
	//escape regex
	glob = escRe.ReplaceAllStringFunc(glob, func(char string) string {
		return `\` + char
	})
	//convert placeholders into regex
	glob = strings.Replace(glob, `GLOB!STAR`, `[^\/]+`, -1)
	glob = strings.Replace(glob, `GLOB!ALL`, `\/([^\/]+\/)*`, -1)
	glob = "^" + glob + "$"
	return regexp.MustCompile(glob)
}

//matcher is configured to match different files
type matcher struct {
	hidden    bool
	include   bool
	str       string
	allFiles  bool
	file, dir *regexp.Regexp
}

func (m *matcher) set(str string) {

	i := strings.LastIndex(str, "/")
	//trailing slash? directory
	if i == len(str)-1 {
		str += "**/*" //implies all files
	}

	m.allFiles = strings.HasSuffix(str, "**/*")

	m.str = str
	m.file = str2regex(str)
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
	//directory match
	if m.dir != nil && !isFile {
		//special case: excluding SOME files?
		//match all dirs since they may contain matches
		if !m.include && !m.allFiles {
			return true
		}
		//directories must have trailing slash
		if !strings.HasSuffix(s, "/") {
			s += "/"
		}
		//include child directories
		if m.include && strings.HasPrefix(m.str, s) {
			return true
		}
		//directory match
		return m.flip(m.dir.MatchString(s))
	}
	return true
}
