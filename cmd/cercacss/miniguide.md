# Cerca CSS Compiler
_a mini guide_

`cercacss` does a few things:
* generates utility classes (design tokens) according to a parameterized css-like config format
* trims generated classes to only the subset that is found in the scanned html folder
* imports and inlines all css files found from the config file
* outputs a single css file, with everything that's needed for the project
## Usage
Typical usage looks something like:

```
./cercacss --html ../../html --out final.css
```

To see all options, run the compiler's help:

```
./cercacss -h 

Usage of ./cercacss:
  -config string
        specify an alternate name of the css config file (default "config.css")
  -css string
        path to folder containing css (including config.css)
  -html string
        path to folder containing html templates
  -out string
        the resulting css file (default "final.css")
```


## `config.css`
* Contains the initial `@import <name>.css;` statements 
	* However, imported files may also import other files and these will be inlined correctly into the final css
* Has the config structure described [elsewhere](https://github.com/cblgh/cerca/issues/29#issuecomment-1025610496) for generating design tokens

### Example config.css
```css
@import "one.css";
@import url("two.css");
@import "three.css";

:colors {
	--main: black;
	--second: red;
}

:scale {
	--01: 0.75;
	--0: 1;
	--1: 1.25;
}

.pad {
	padding-left: var(--scale)rem;
	padding-right: var(--scale)rem;
}

.text {
	color: var(--colors);
}
``` 

Which, if all tokens are used in the html, will output:
```css
/* the generated utility classes */
.pad-0 {
         padding-left: 1rem;
         padding-right: 1rem;
}
.pad-1 {
         padding-left: 1.25rem;
         padding-right: 1.25rem;
}
.pad-01 {
         padding-left: 0.75rem;
         padding-right: 0.75rem;
}
.text-main {
         color: black;
}
.text-second {
         color: red;
}
/* the imported css statements */
body {
    margin-top: 2rem;
    max-width: 50rem;
}

header {
    background-color: orange;
}

footer {
    background-color: wheat;
}

main {
    padding-left: 2rem;
    padding-right: 2rem;
}
```
