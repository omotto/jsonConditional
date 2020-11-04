package jsonconditional

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseCondition(t *testing.T) {
	p := New(IBM)
	query, err := p.Parse(`{"$or":[{"Gestiones.idComercial":{"$lte":300}},{"Gestiones.texto":""},{"$and":[{"tblTiposGestion.borrado":0},{"Gestiones.idtipocheckin":{"$nin":[4,5]}}]}]}`)
	if err != nil {
		t.Fatal(err)
	} else {
		assert.Equal(t, query, ` ( Gestiones.idComercial  <= 300  OR  Gestiones.texto  = ''  OR  ( tblTiposGestion.borrado  = 0  AND  Gestiones.idtipocheckin  NOT IN (4,5) )  ) `)
	}
	p = New(JSONLOGIC)
	query, err = p.Parse(`{"and": [ { "and": [ { "<=": [ { "var": "Gestiones.idComercial" }, 300 ] }, { ">=": [ { "var": "Gestiones.texto" }, "2006-01-02 23:08:30" ] } ] }, { "or": [ { "==": [ { "var": "tblTiposGestion.borrado" }, "0" ] }, { "in": [ { "var": "Gestiones.idtipocheckin" }, [4, 5] ] } ] } ] }`)
	if err != nil {
		t.Fatal(err)
	} else {
		assert.Equal(t, query, ` ( (  Gestiones.idComercial  <=  300   AND   Gestiones.texto  >=  CONVERT(DATETIME, '2006-01-02 23:08:30')   )  AND  (  tblTiposGestion.borrado  =  '0'   OR   Gestiones.idtipocheckin  IN   (4,5)   )  ) `)
	}
}
