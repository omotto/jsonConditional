[![GoDoc](http://godoc.org/github.com/omotto/jsonConditional?status.png)](http://godoc.org/github.com/omotto/jsonConditional)
[![Build Status](https://travis-ci.com/omotto/jsonConditional.svg?branch=main)](https://travis-ci.com/omotto/jsonConditional)
[![Coverage Status](https://coveralls.io/repos/github/omotto/jsonConditional/badge.svg)](https://coveralls.io/github/omotto/jsonConditional)

# jsonConditional

Parse JSON condition formats in MS SQL query format
 
```
// New JSON Parser
p := New(IBM) // Implemented IBM and JSONLOGIC parsers
// Parse and check
if query, err := p.Parse(`{"$or":[{"Gestiones.idComercial":{"$lte":300}},{"Gestiones.texto":""},{"$and":[{"tblTiposGestion.borrado":0},{"Gestiones.idtipocheckin":{"$nin":[4,5]}}]}]}`); err == nil {
    fmt.println(query)
}
```

