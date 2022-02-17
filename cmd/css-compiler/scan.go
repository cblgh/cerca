package main

import (
  "fmt"
  "os"
  "regexp"
  "cerca/util"
  "strings"
)

var classPattern = regexp.MustCompile(`class="(.*?)"`)
func main () {
  if len(os.Args) < 2 {
    fmt.Println("usage: scan <html file to scan>")
    os.Exit(0)
  }

  data, err := os.ReadFile(os.Args[1])
  util.Check(err, "read template file")
  html := string(data)

  var classList []string
  classMap := make(map[string]bool)
  for _, line := range strings.Split(html, "\n") {
    matches := classPattern.FindStringSubmatch(line)
    if len(matches) > 0 {
        for _, match := range strings.Fields(matches[1]) {
          if !classMap[match] {
          classList = append(classList, match)
          classMap[match] = true
        }
      }
    }
  }

  for _, m := range classList {
    fmt.Println(m)
  }
}
