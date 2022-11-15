# microformats

[![GoDoc](https://img.shields.io/badge/godoc-reference-blue)](https://pkg.go.dev/willnorris.com/go/microformats)
[![Test Status](https://github.com/willnorris/microformats/workflows/ci/badge.svg)](https://github.com/willnorris/microformats/actions?query=workflow%3Aci)
[![Test Coverage](https://codecov.io/gh/willnorris/microformats/branch/main/graph/badge.svg)](https://codecov.io/gh/willnorris/microformats)

microformats is a go library and tool for parsing [microformats][], supporting both classic v1 and [v2 syntax][].
It is based on Andy Leap's [original library][andyleap/microformats].

[microformats]: https://microformats.io/
[v2 syntax]: https://microformats.org/wiki/microformats-2
[andyleap/microformats]: https://github.com/andyleap/microformats

## Usage

To see this package in action, the simplest way is to install the command line
app and use it to fetch and parse a webpage with microformats on it:

``` sh
% go install willnorris.com/go/microformats/cmd/gomf@latest
% gomf https://indieweb.org
```

To use it in your own code, import the package:

``` go
import "willnorris.com/go/microformats"
```

If you have the HTML contents of a page in an [io.Reader][], call [Parse][] like in this example:

``` go
content := `<article class="h-entry"><h1 class="p-name">Hello</h1></article>`
r := strings.NewReader(content)

data := microformats.Parse(r, nil)

// do something with data, or just print it out as JSON:
enc := json.NewEncoder(os.Stdout)
enc.SetIndent("", "  ")
enc.Encode(data)
```

Alternately, if you have already parsed the page and have an [html.Node][], then call [ParseNode][].
For example, you might want to select a subset of the DOM, and parse only that for microformats.
An example of doing this with the [goquery package] can be seen in [cmd/gomf/main.go](cmd/gomf/main.go).

To see that in action using the gomf app installed above,
you can parse the microformats from indieweb.org that appear within the `#content` element:

``` sh
% gomf https://indieweb.org "#content"

{
  "items": [
    {
      "id": "content",
      "type": [
        "h-entry"
      ],
      "properties": ...
      "children": ...
    }
  ],
  "rels": {},
  "rel-urls": {}
}
```

[Parse]: https://pkg.go.dev/willnorris.com/go/microformats#Parse
[ParseNode]: https://pkg.go.dev/willnorris.com/go/microformats#ParseNode
[io.Reader]: https://golang.org/pkg/io/#Reader
[html.Node]: https://pkg.go.dev/golang.org/x/net/html#Node
[goquery package]: https://github.com/PuerkitoBio/goquery
