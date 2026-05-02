package query

// ListOptions represents common pagination options for list queries
// Used across namespace, module, and other list operations
// Python reference: terrareg/models.py get_all() offset/limit parameters
type ListOptions struct {
	Offset int
	Limit  int
}

// IsPaginated returns true if pagination is enabled (limit > 0)
func (o *ListOptions) IsPaginated() bool {
	return o != nil && o.Limit > 0
}
