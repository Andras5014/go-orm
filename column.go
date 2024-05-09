package go_orm

type Column struct {
	name  string
	alias string
}

func C(name string) Column {
	return Column{name: name}
}

func (c Column) Eq(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opEq,
		right: valueOf(arg),
	}
}
func (c Column) assign() {

}
func (c Column) As(alias string) Column {
	return Column{
		name:  c.name,
		alias: alias,
	}
}

func valueOf(arg any) Expression {
	switch val := arg.(type) {
	case Expression:
		return val
	default:
		return value{
			arg: val,
		}
	}
}
func (Column) expr() {}
func (c Column) selectable() {

}
