package main

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var moduleRgx = regexp.MustCompile(`(?m)module\s+(.+)`)
var importRgx = regexp.MustCompile(`(?m)(import\s*\((.|\s)*?\))`)
var module string     // e.g. gitlab.com/services/tinkoff
var repository string // e.g. gitlab.com

func main() {
	var path string
	var err error
	if len(os.Args[1:]) == 0 || os.Args[1] == "." {
		path, err = os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		path = os.Args[1]
	}

	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		log.Fatal(err)
	}

	sortImports(path)
}

func sortImports(path string) {
	files, err := os.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	if len(module) == 0 {
		for _, file := range files {
			if file.IsDir() || filepath.Ext(file.Name()) != ".mod" {
				continue
			}
			buf, err := os.ReadFile(filepath.Join(path, file.Name()))
			if err != nil {
				log.Fatal(err)
			}
			content := string(buf)
			res := moduleRgx.FindAllStringSubmatch(content, -1)
			module = res[0][1]
		}
		if len(module) == 0 {
			log.Fatal(errors.New("not a go directory"))
		}
		split := strings.Split(module, "/")
		repository = split[0]
	}

	for _, file := range files {
		if file.IsDir() {
			sortImports(filepath.Join(path, file.Name()))
			continue
		}

		if filepath.Ext(file.Name()) != ".go" {
			continue
		}

		filePath := filepath.Join(path, file.Name())
		buf, err := os.ReadFile(filePath)
		if err != nil {
			log.Fatal(err)
		}
		content := string(buf)
		res := importRgx.FindAllStringSubmatch(content, -1)
		if len(res) == 0 {
			continue
		}

		importsMap := map[string][]string{
			"std":          {},
			"ourInternal":  {},
			"ourExternal":  {},
			"libsExternal": {},
		}

		importBlock := res[0][0]
		importBlock = strings.Replace(importBlock, "import (", "", 1)
		importBlock = strings.Replace(importBlock, ")", "", 1)
		importBlock = strings.ReplaceAll(importBlock, "\t", "")
		split := strings.Split(importBlock, "\n")
		for _, imp := range split {
			imp2 := strings.Trim(imp, " ")
			imp2 = strings.Trim(imp, "\t")
			if len(imp2) == 0 {
				continue
			}

			if !strings.Contains(imp, ".") {
				importsMap["std"] = append(importsMap["std"], imp)
			} else if strings.Contains(imp, module) {
				importsMap["ourInternal"] = append(importsMap["ourInternal"], imp)
			} else if !strings.Contains(imp, module) && strings.Contains(imp, repository) {
				importsMap["ourExternal"] = append(importsMap["ourExternal"], imp)
			} else {
				importsMap["libsExternal"] = append(importsMap["libsExternal"], imp)
			}
		}

		var substitution strings.Builder
		substitution.WriteString("import (")
		writeImportPart(importsMap["std"], &substitution)
		writeImportPart(importsMap["ourInternal"], &substitution)
		writeImportPart(importsMap["ourExternal"], &substitution)
		writeImportPart(importsMap["libsExternal"], &substitution)
		substitution.WriteString(")")

		sortedImports := importRgx.ReplaceAllString(content, substitution.String())
		err = os.WriteFile(filePath, []byte(sortedImports), os.ModeExclusive)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func writeImportPart(importsMap []string, substitution *strings.Builder) {
	if len(importsMap) > 0 {
		substitution.WriteString("\n")
		for _, v := range importsMap {
			substitution.WriteString("\t" + v + "\n")
		}
	}
}
