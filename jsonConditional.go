package jsonconditional

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ConditionalType represents the accepted conditional JSON types
type ConditionalType int

// LayoutISO is the current accepted DATE TIME format
const LayoutISO = "2006-01-02 15:04:05"

const (
	// IBM conditional format
	IBM ConditionalType = iota
	// JSONLOGIC conditional format
	JSONLOGIC
)

// ConditionalParser is the Interface that must be followed by the different parsers
type ConditionalParser interface {
	Parse(jsonLogic string) (query string, err error)
}

// IBMCondition https://www.ibm.com/support/knowledgecenter/SSEPEK_11.0.0/json/src/tpc/db2z_jsonqueryoperator.html
type IBMCondition struct {
}

// JSONLogicCondition http://jsonlogic.com/operations.html
type JSONLogicCondition struct {
}

// New Factory of JSON Conditional parsers
func New(t ConditionalType) ConditionalParser {
	if t == IBM {
		return IBMCondition{}
	}
	return JSONLogicCondition{}
}

// -- IBM to SQL
func setValue(condition string, value interface{}) (ret string) {
	if _, ok := value.(string); ok {
		ret = condition + fmt.Sprintf("'%v' ", value)
	} else {
		ret += condition + fmt.Sprintf("%v ", value)
	}
	return ret
}

func parseValue(key string, value interface{}) (query string, err error) {
	if key[0] == '$' {
		switch strings.ToLower(key) {
		case "$and":
			if keyValues, ok := value.([]interface{}); ok {
				var subQuery string
				for _, keyValue := range keyValues {
					for k, v := range keyValue.(map[string]interface{}) {
						if str, e := parseValue(k, v); e == nil {
							subQuery += str + " AND "
						} else {
							return "", e
						}
					}
				}
				query += fmt.Sprintf(" (%s) ", subQuery[:len(subQuery)-4])
			} else {
				err = errors.New("NOT condition bad formatted")
			}
		case "$or":
			if keyValues, ok := value.([]interface{}); ok {
				var subQuery string
				for _, keyValue := range keyValues {
					for k, v := range keyValue.(map[string]interface{}) {
						if str, e := parseValue(k, v); e == nil {
							subQuery += str + " OR "
						} else {
							return "", e
						}
					}
				}
				query += fmt.Sprintf(" (%s) ", subQuery[:len(subQuery)-3])
			} else {
				err = errors.New("NOT condition bad formatted")
			}
		case "$not":
			if keyValue, ok := value.(map[string]interface{}); ok {
				var subQuery string
				for k, v := range keyValue {
					if str, e := parseValue(k, v); e == nil {
						subQuery += str
					} else {
						err = e
						break
					}
				}
				query += fmt.Sprintf(" NOT (%s)", subQuery)
			} else {
				err = errors.New("NOT condition bad formatted")
			}
		case "$nor":
			err = errors.New("NOR condition not implemented")
		case "$ne":
			query += fmt.Sprintf(" != %v", value)
		case "$in":
			if keyValues, ok := value.([]interface{}); ok && len(keyValues) > 0 {
				if _, ok := keyValues[0].(string); ok {
					query += fmt.Sprintf(" IN ('%v')", strings.Trim(strings.Replace(fmt.Sprint(value), " ", "','", -1), "[]"))
				} else if _, ok := keyValues[0].(float64); ok {
					query += fmt.Sprintf(" IN (%v)", strings.Trim(strings.Replace(fmt.Sprint(value), " ", ",", -1), "[]"))
				}
			} else {
				err = errors.New("IN condition bad formatted")
			}
		case "$nin":
			if keyValues, ok := value.([]interface{}); ok && len(keyValues) > 0 {
				if _, ok := keyValues[0].(string); ok {
					query += fmt.Sprintf(" NOT IN ('%v')", strings.Trim(strings.Replace(fmt.Sprint(value), " ", "','", -1), "[]"))
				} else if _, ok := keyValues[0].(float64); ok {
					query += fmt.Sprintf(" NOT IN (%v)", strings.Trim(strings.Replace(fmt.Sprint(value), " ", ",", -1), "[]"))
				}
			} else {
				err = errors.New("NOT IN condition bad formatted")
			}
		case "$lte":
			query += setValue(" <= " , value)
		case "$gte":
			query += setValue(" >= ", value)
		case "$gt":
			query += setValue(" > ", value)
		case "$lt":
			query += setValue(" < ", value)
		}
	} else {
		query += fmt.Sprintf(" %v ", key)
		if keyValue, ok := value.(map[string]interface{}); ok {
			for k, v := range keyValue {
				if str, e := parseValue(k, v); e == nil {
					query += str
				} else {
					err = e
					break
				}
			}
		} else {
			query += setValue(" = ", value)
		}

	}
	return query, err
}

// Parse input IBM conditional format to SQL conditional query
func (i IBMCondition) Parse(jsonLogic string) (query string, err error) {
	var raw map[string]interface{}
	err = json.Unmarshal([]byte(jsonLogic), &raw)
	for key, value := range raw {
		if str, e := parseValue(key, value); e == nil {
			query += str
		} else {
			err = e
			break
		}
	}
	return query, err
}

// -- JSONLogic to SQL
func setValueFormat(value interface{}) (ret string) {
	if _, ok := value.(string); ok {
		if _, err := time.Parse(LayoutISO, value.(string)); err == nil {
			ret = fmt.Sprintf(" CONVERT(DATETIME, '%s') ", value.(string))
		} else {
			ret = fmt.Sprintf(" '%s' ", strings.Replace(value.(string), "'", "''", -1))
		}
	} else {
		if keyValues, ok := value.([]interface{}); ok && len(keyValues) > 0 {
			if _, ok := keyValues[0].(string); ok {
				ret = fmt.Sprintf(" ('%v') ", strings.Trim(strings.Replace(fmt.Sprint(value), " ", "','", -1), "[]"))
			} else {
				ret = fmt.Sprintf("  (%v) ", strings.Trim(strings.Replace(fmt.Sprint(value), " ", ",", -1), "[]"))
			}
		} else {
			ret = fmt.Sprintf(" %v ", value)
		}
	}
	return ret
}

func parseJSONLogicValue(key string, value interface{}) (query string, err error) {
	switch strings.ToLower(key) {
		case "and":
			if keyValues, ok := value.([]interface{}); ok {
				var subQuery string
				for _, keyValue := range keyValues {
					for k, v := range keyValue.(map[string]interface{}) {
						if str, e := parseJSONLogicValue(k, v); e == nil {
							subQuery += str + " AND "
						} else {
							return "", e
						}
					}
				}
				query += fmt.Sprintf(" (%s) ", subQuery[:len(subQuery)-4])
			} else {
				err = errors.New("AND condition bad formatted")
			}
		case "or":
			if keyValues, ok := value.([]interface{}); ok {
				var subQuery string
				for _, keyValue := range keyValues {
					for k, v := range keyValue.(map[string]interface{}) {
						if str, e := parseJSONLogicValue(k, v); e == nil {
							subQuery += str + " OR "
						} else {
							return "", e
						}
					}
				}
				query += fmt.Sprintf(" (%s) ", subQuery[:len(subQuery)-3])
			} else {
				err = errors.New("OR condition bad formatted")
			}
		case "var":
			query += fmt.Sprintf(" %s ", value)
		case "==":
			if keyValues, ok := value.([]interface{}); ok {
				var subQuery string
				for _, keyValue := range keyValues {
					if _, ok := keyValue.(map[string]interface{}); ok {
						for k, v := range keyValue.(map[string]interface{}) {
							if str, e := parseJSONLogicValue(k, v); e == nil {
								subQuery += str + " = "
							} else {
								return "", e
							}
						}
					} else {
						subQuery += setValueFormat(keyValue)
					}
				}
				query += fmt.Sprintf(" %s ", subQuery)
			} else {
				err = errors.New("EQ condition bad formatted")
			}
		case "!=":
			if keyValues, ok := value.([]interface{}); ok {
				var subQuery string
				for _, keyValue := range keyValues {
					if _, ok := keyValue.(map[string]interface{}); ok {
						for k, v := range keyValue.(map[string]interface{}) {
							if str, e := parseJSONLogicValue(k, v); e == nil {
								subQuery += str + " <> "
							} else {
								return "", e
							}
						}
					} else {
						subQuery += setValueFormat(keyValue)
					}
				}
				query += fmt.Sprintf(" %s ", subQuery)
			} else {
				err = errors.New("NEQ condition bad formatted")
			}
		case ">":
			if keyValues, ok := value.([]interface{}); ok {
				var subQuery string
				for _, keyValue := range keyValues {
					if _, ok := keyValue.(map[string]interface{}); ok {
						for k, v := range keyValue.(map[string]interface{}) {
							if str, e := parseJSONLogicValue(k, v); e == nil {
								subQuery += str + " > "
							} else {
								return "", e
							}
						}
					} else {
						subQuery += setValueFormat(keyValue)
					}
				}
				query += fmt.Sprintf(" %s ", subQuery)
			} else {
				err = errors.New("GT condition bad formatted")
			}
		case ">=":
			if keyValues, ok := value.([]interface{}); ok {
				var subQuery string
				for _, keyValue := range keyValues {
					if _, ok := keyValue.(map[string]interface{}); ok {
						for k, v := range keyValue.(map[string]interface{}) {
							if str, e := parseJSONLogicValue(k, v); e == nil {
								subQuery += str + " >= "
							} else {
								return "", e
							}
						}
					} else {
						subQuery += setValueFormat(keyValue)
					}
				}
				query += fmt.Sprintf(" %s ", subQuery)
			} else {
				err = errors.New("GTE condition bad formatted")
			}
		case "<":
			if keyValues, ok := value.([]interface{}); ok {
				var subQuery string
				for _, keyValue := range keyValues {
					if _, ok := keyValue.(map[string]interface{}); ok {
						for k, v := range keyValue.(map[string]interface{}) {
							if str, e := parseJSONLogicValue(k, v); e == nil {
								subQuery += str + " < "
							} else {
								return "", e
							}
						}
					} else {
						subQuery += setValueFormat(keyValue)
					}
				}
				query += fmt.Sprintf(" %s ", subQuery)
			} else {
				err = errors.New("LT condition bad formatted")
			}
		case "<=":
			if keyValues, ok := value.([]interface{}); ok {
				var subQuery string
				for _, keyValue := range keyValues {
					if _, ok := keyValue.(map[string]interface{}); ok {
						for k, v := range keyValue.(map[string]interface{}) {
							if str, e := parseJSONLogicValue(k, v); e == nil {
								subQuery += str + " <= "
							} else {
								return "", e
							}
						}
					} else {
						subQuery += setValueFormat(keyValue)
					}
				}
				query += fmt.Sprintf(" %s ", subQuery)
			} else {
				err = errors.New("LTE condition bad formatted")
			}
		case "in":
			if keyValues, ok := value.([]interface{}); ok {
				var subQuery string
				for _, keyValue := range keyValues {
					if _, ok := keyValue.(map[string]interface{}); ok {
						for k, v := range keyValue.(map[string]interface{}) {
							if str, e := parseJSONLogicValue(k, v); e == nil {
								subQuery += str + " IN "
							} else {
								return "", e
							}
						}
					} else {
						subQuery += setValueFormat(keyValue)
					}
				}
				query += fmt.Sprintf(" %s ", subQuery)
			} else {
				err = errors.New("IN condition bad formatted")
			}
		case "!":
			if keyValues, ok := value.([]interface{}); ok {
				var subQuery string
				for _, keyValue := range keyValues {
					if _, ok := keyValue.(map[string]interface{}); ok {
						for k, v := range keyValue.(map[string]interface{}) {
							if str, e := parseJSONLogicValue(k, v); e == nil {
								subQuery += str + " NOT IN "
							} else {
								return "", e
							}
						}
					} else {
						subQuery += setValueFormat(keyValue)
					}
				}
				query += fmt.Sprintf(" %s ", subQuery)
			} else {
				err = errors.New("NOT IN condition bad formatted")
			}
		default:
			err = errors.New(key + " operator not supported")
	}
	return query, err
}

// Parse input JSON Logic format to SQL conditional query
func (j JSONLogicCondition) Parse(jsonLogic string) (query string, err error) {
	var raw map[string]interface{}
	err = json.Unmarshal([]byte(jsonLogic), &raw)
	for key, value := range raw {
		if str, e := parseJSONLogicValue(key, value); e == nil {
			query += str
		} else {
			err = e
			break
		}
	}
	return query, err
}