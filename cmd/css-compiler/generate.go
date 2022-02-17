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
  // keep generated class names in a list to keep order stable across executions
  classNames := make([]string, 0)
  classDeclarations := make(map[string][]string)
  for _, line := range lines {
    // empty
    if strings.HasPrefix(line, "{") {
      continue
    }
    // reset block
    if strings.HasPrefix(line, "}") {
      if insideClass {
        for _, id := range classNames {
          fmt.Printf("%s {\n", id)
          for _, declaration := range classDeclarations[id] {
            fmt.Println("\t", declaration)
          }
          fmt.Println("}")
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
}
