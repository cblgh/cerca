package main

import (
  "fmt"
  "strings"
  "sort"
  "path/filepath"
  "os"
  "regexp"
  "cerca/util"
  "flag"
)

/*
files and folders:
the css configuration file
the import file
the location of the resulting generated file
the location of the final, combined css file
the location of the css files
the location of the html templates

one solution: only have path to css; 
  put config stuff into config.css
  put imports at the top of config.css (read and hold the imports separately from the generated tokens though!)
  output style.css
*/
func main () {
  var htmlPath string
  var cssPath string
  var outputFile string
  flag.StringVar(&cssPath, "css", "", "path to folder containing css (include config.css)")
  flag.StringVar(&htmlPath, "html", "", "path to folder containing html templates")
  flag.StringVar(&outputFile, "out", "final.css", "the resulting css file")

  flag.Parse()

  // read the custom css config, used to generate our design tokens
  b, err := os.ReadFile(filepath.Join(cssPath, "config.css"))
  util.Check(err, "read config.css")
  lines := strings.Split(string(b), "\n")

  designTokens := generateDesignTokens(lines)

  html := readHTML(htmlPath)
  foundTokens, tokenMap := getTokensFromHTML(html)
  fmt.Println("found tokens", len(foundTokens), foundTokens)

  tokenMap["pad-01"] = true
  tokenMap["text-main"] = true

  filteredGeneratedCSS := filterDesignTokens(strings.Join(designTokens, "\n"), tokenMap)

  importedCSS := readImportedCSS(cssPath, lines)
  outputCSS(append(filteredGeneratedCSS, importedCSS...)...)
}

// read a custom css-based markup & generate css classes (aka design tokens) using the markup's scales, values, and classes
func generateDesignTokens (lines []string) []string {
  var tokens []string
  vars := make(map[string]map[string]string)
  var insideClass bool
  var insideVars bool
  var identifier string
  // keep generated class names in a list to keep order stable across executions
  classNames := make([]string, 0)
  classDeclarations := make(map[string][]string)
  for _, line := range lines {
    // skip: it's either start of a block or an import
    if strings.HasPrefix(line, "{") || strings.HasPrefix(line, "@") {
      continue
    }
    // reset block
    if strings.HasPrefix(line, "}") {
      if insideClass {
        for _, id := range classNames {
          tokens = append(tokens, fmt.Sprintf("%s {\n", id))
          for _, declaration := range classDeclarations[id] {
            tokens = append(tokens, fmt.Sprintln("\t", declaration))
          }
          tokens = append(tokens, fmt.Sprintln("}"))
        }
        classDeclarations = make(map[string][]string)
        classNames = make([]string, 0)
      }
      insideClass = false
      insideVars = false
    }

    // inside a css variable block
    if insideVars {
      parts := strings.Split(strings.TrimSpace(line), ":")
      name := strings.TrimPrefix(parts[0], "--")
      val := strings.TrimSuffix(parts[1], ";")
      vars[identifier][name] = strings.TrimSpace(val)
    }

    // declare new css variable
    if strings.HasPrefix(line, ":") {
      insideVars = true
      identifier = strings.TrimSpace(strings.TrimSuffix(line[1:], "{"))
      if _, exists := vars[identifier]; !exists {
        vars[identifier] = make(map[string]string)
      }
      continue
    }

    // inside a class block
    if insideClass {
      parts := strings.Split(strings.TrimSpace(line), ":")
      property := parts[0]
      varNamePattern := regexp.MustCompile(`.+--(\w+)\)(\w*);`)
      matches := varNamePattern.FindStringSubmatch(parts[1])
      if len(matches) == 0 {
        continue
      }
      varName := matches[1]
      unit := matches[2]
      for key, val := range vars[varName] {
        id := fmt.Sprintf(".%s-%s", identifier, key)
        declaration := fmt.Sprintf("%s: %s%s;", property, val, unit)
        if _, exist := classDeclarations[id]; !exist {
          classNames = append(classNames, id)
          classDeclarations[id] = make([]string, 0)
        }
        classDeclarations[id] = append(classDeclarations[id], declaration)
      }
    }

    // declare new css class
    if strings.HasPrefix(line, ".") {
      insideClass = true
      identifier = strings.TrimSpace(strings.TrimSuffix(line[1:], "{"))
      continue
    }
  }

  return tokens
}

// read all html files at destination folder, and return the string contents of each file
func readHTML (dir string) []string {
  ed := util.Describe("read html")
  files, err := os.ReadDir(dir)
  ed.Check(err, "list subfolders")
  var html []string
  // TODO(2022-03-03): recursively descend into subsubfolders
  for _, file := range files {
    if filepath.Ext(file.Name()) != ".html" {
      continue
    }
    filepath := filepath.Join(dir, file.Name())
    b, err := os.ReadFile(filepath)
    ed.Check(err, "read %s", file)
    html = append(html, string(b))
  }
  return html
}

// get tokens (classes and html tags) from the passed in html
func getTokensFromHTML (html []string) ([]string, map[string]bool) {
  var tokens []string
  tokenMap := make(map[string]bool)
  classPattern := regexp.MustCompile(`class="(.*?)"`)
  htmlTagPattern := regexp.MustCompile(`<(\w+)`)
  // parse out all tokens we might be interested in wrt css-styling 
  for _, contents := range html {
    // find all classes used in the html file
    classMatches := classPattern.FindAllStringSubmatch(contents, -1)
    for _, matches := range classMatches {
      if len(matches) >= 2 {
        classes := strings.Fields(matches[1])
        for _, token := range classes {
          tokenMap[token] = true
        }
      }
    }
    // find all html tag tokens referenced in file
    tagMatches := htmlTagPattern.FindAllStringSubmatch(contents, -1)
    for _, matches := range tagMatches {
      if len(matches) >= 2 {
        tokenMap[matches[1]] = true
      }
    }
  }
  skipTags := []string{"style", "title", "meta", "head"}
  var skip bool
  // deduplicate the found tokens using the token map
  for tag := range tokenMap {
    for _, skipTag := range skipTags {
      if tag == skipTag {
        tokenMap[tag] = false
        skip = true
        break
      }
    }
    if !skip {
      tokens = append(tokens, tag)
    }
    skip = false
  }
  // before returning: stably sort the tokens, making generated result consistent across runs 
  // (might be a boon in debugging in future)
  sort.Strings(tokens)
  return tokens, tokenMap
}

// filter the generated design tokens based on the HTML's used tokens
func filterDesignTokens (generated string, used map[string]bool) []string {
  var usefulCSS []string
  var classPattern = regexp.MustCompile(`\.(\S*)\s*{`)
  // inside a { css } block
  var insideBlock bool
  // inside a css block which contains lines we want to save
  var insideUsefulBlock bool
  for _, line := range strings.Split(generated, "\n") {
    content := strings.TrimSpace(line)
    if len(content) == 0 || strings.HasPrefix(content, "{") {
      continue
    }
    if content == "}" {
      insideBlock = false
    }
    if !insideBlock {
      matches := classPattern.FindStringSubmatch(content)
      if len(matches) >= 2 {
        insideBlock = true
        token := matches[1]
        insideUsefulBlock = used[token]
      }
    }
    if insideUsefulBlock {
      usefulCSS = append(usefulCSS, line)
    }
  }
  return usefulCSS
}

// match: @import "navigation.css"; 
var importPattern = regexp.MustCompile(`@import\s*"(\S+)";`)
// match: @import url("navigation.css"); 
var altImportPattern = regexp.MustCompile(`@import\s*url\("(\S+)"\);`)

func collectImports(lines []string) []string {
  paths := make([]string, 0)
  for _, line := range lines {
    if strings.HasPrefix(line, "@import") {
      matches := importPattern.FindStringSubmatch(line)
      if len(matches) == 0 {
        matches = altImportPattern.FindStringSubmatch(line)
      }
      if len(matches) > 1 {
        paths = append(paths, matches[1])
      }
    }
  }
  return paths
}

func readImportedCSS (cssPath string, lines []string) []string {
  var css []string
  for _, filename := range collectImports(lines) {
    b, err := os.ReadFile(filepath.Join(cssPath, filename))
    util.Check(err, "read @imported file")
    css = append(css, string(b))
  }
  return css
}

// output the final css: the result of token generation, imports, filtering generated tokens by what are used
func outputCSS (css ...string) {
  for _, line := range css {
    fmt.Println(line)
  }
}
