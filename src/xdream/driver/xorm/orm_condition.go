package xorm

import (
	"fmt"
	"strings"
)

const (
	ORDER_DESC = "DESC"
	ORDER_ASC  = "ASC"
)

type Pager struct {
	Page  int
	Count int
}

func (this *Pager) GetLimit() (int, int) {
	return (this.Page - 1) * this.Count, this.Count
}

type Condition struct {
	fields string
	conds  string
	values []interface{}
	limit  string
	order  string
}

func NewCondition() *Condition {
	cond := &Condition{}
	cond.values = []interface{}{}
	return cond
}

func (this *Condition) Fields(fields string) *Condition {
	this.fields = strings.ToLower(fields)
	return this
}

func (this *Condition) And(col string, v interface{}, args ...string) *Condition {
	oper := "="
	if len(args) > 0 {
		oper = args[0]
	}
	return this.Append(col, v, oper, "AND")
}

func (this *Condition) Or(col string, v interface{}, args ...string) *Condition {
	oper := "="
	if len(args) > 0 {
		oper = args[0]
	}
	return this.Append(col, v, oper, "OR")
}

func (this *Condition) Limit(start int, end ...int) *Condition {
	if len(end) == 0 {
		this.limit = fmt.Sprintf("%d", start)
	} else {
		this.limit = fmt.Sprintf("%d,%d", start, end[0])
	}
	return this
}

func (this *Condition) Page(pager *Pager) *Condition {
	s, e := pager.GetLimit()
	this.limit = fmt.Sprintf("%d,%d", s, e)
	return this
}

func (this *Condition) OrderBy(col string, order string) *Condition {
	if order == "" {
		order = "DESC"
	}
	if len(this.order) > 0 {
		this.order = fmt.Sprintf("%s,`%s` %s", this.order, col, order)
	} else {
		this.order = fmt.Sprintf("`%s` %s", col, order)
	}
	return this
}

func (this *Condition) Append(col string, v interface{}, oper string, typs ...string) *Condition {
	typ := ""
	if len(this.conds) == 0 {
		typ = ""
	} else {
		if len(typs) > 0 {
			typ = typs[0]
		} else {
			typ = "AND"
		}
	}

	this.conds = fmt.Sprintf("%s %s `%s` %s ?", this.conds, typ, col, oper)
	this.values = append(this.values, v)
	return this
}

func (this *Condition) In(col string, values ...interface{}) *Condition {
	return this.AppendIn(col, values, "AND")
}

func (this *Condition) NotIn(col string, values ...interface{}) *Condition {
	return this.AppendNotIn(col, values, "AND")
}

func (this *Condition) AppendIn(col string, values []interface{}, typ string) *Condition {
	if len(values) == 0 {
		return this
	}
	if len(this.conds) == 0 {
		typ = ""
	}

	holderStr := strings.Repeat("?,", len(values))
	holderStr = holderStr[0 : len(holderStr)-1]
	this.conds = fmt.Sprintf("%s %s `%s` IN (%s)", this.conds, typ, col, holderStr)
	this.values = append(this.values, values...)

	return this
}

func (this *Condition) AppendNotIn(col string, values []interface{}, typ string) *Condition {
	if len(values) == 0 {
		return this
	}
	if len(this.conds) == 0 {
		typ = ""
	}

	holderStr := strings.Repeat("?,", len(values))
	holderStr = holderStr[0 : len(holderStr)-1]
	this.conds = fmt.Sprintf("%s %s `%s` NOT IN (%s)", this.conds, typ, col, holderStr)
	this.values = append(this.values, values...)

	return this
}

func (this *Condition) GetCondition() (string, []interface{}) {
	if len(this.conds) > 0 {
		this.conds = fmt.Sprintf("WHERE %s", this.conds)
	}

	if len(this.order) > 0 {
		this.conds = fmt.Sprintf("%s ORDER BY %s", this.conds, this.order)
	}

	if len(this.limit) > 0 {
		this.conds = fmt.Sprintf("%s LIMIT %s", this.conds, this.limit)
	}

	return this.conds, this.values
}

func (this *Condition) GetFields() string {
	if len(this.fields) == 0 {
		return "*"
	}
	return this.fields
}
