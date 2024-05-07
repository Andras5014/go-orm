package go_orm

type Column struct {
	name string
}

func (Column) expr() {}
func (c Column) selectable() {

}
func C(name string) Column {
	return Column{name: name}
}

func (c Column) Eq(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opEq,
		right: value{arg: arg},
	}
}
