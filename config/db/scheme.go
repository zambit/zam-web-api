package db

// Scheme of db connection configuration
type Scheme struct {
	// URI contains all necessary connection parts in URI form
	// Described here https://www.postgresql.org/docs/current/static/libpq-connect.html#id-1.7.3.8.3.2.
	URI string
}
