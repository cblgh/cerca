package main

import (
  "os"
  "fmt"
  "strings"
  "regexp"
  "cerca/util"
)

func main () {
  b, err := os.ReadFile("./import.css")
  util.Check(err, "read import.css")

  lines := string(b)
  imports := collectImports(lines)
  for _, filepath := range imports {
    b, err := os.ReadFile(filepath)
    util.Check(err, "read @imported file")
    fmt.Println(string(b))
  }
}

// matches: @import "navigation.css"; 
var importPattern = regexp.MustCompile(`@import\s*"(\S+)";`)
// matches: @import url("navigation.css"); 
var altImportPattern = regexp.MustCompile(`@import\s*url\("(\S+)"\);`)
func collectImports(input string) []string {
  paths := make([]string, 0)
  for _, line := range strings.Split(input, "\n") {
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
