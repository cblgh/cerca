package main

import (
  "fmt"
  "os"
  "cerca/util"
  "strings"
  "regexp"
)

func main () {
  b, err := os.ReadFile("./example.css")
  util.Check(err, "read example css")
  lines := strings.Split(string(b), "\n")
  vars := make(map[string]map[string]string)
  var insideClass bool
  var insideVars bool
  var identifier string
  for _, line := range lines {
    // empty
    if strings.HasPrefix(line, "{") {
      continue
    }
    // reset block
    if strings.HasPrefix(line, "}") {
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
        fmt.Printf(".%s-%s { %s: %s%s; }\n", identifier, key, property, val, unit)
      }
    }
    // declare new css class
    if strings.HasPrefix(line, ".") {
      insideClass = true
      identifier = strings.TrimSpace(strings.TrimSuffix(line[1:], "{"))
      continue
    }
  }
}
