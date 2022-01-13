package index

type Index interface {
	Index(docs ...interface{}) error
}
