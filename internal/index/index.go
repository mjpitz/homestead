package index

type Config struct {
	Endpoint string `json:"endpoint" usage:"a dsn string pointing to the database where we will write our data"`
}

type Index interface {
	Index(docs ...interface{}) error
}
