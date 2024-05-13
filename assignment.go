package go_orm

type Assignment struct {
	col string
	val any
}

func (a Assignment) expr() {}

func (a Assignment) assign() {
}
func Assign(col string, val any) Assignment {
	return Assignment{
		col: col,
		val: val,
	}
}
