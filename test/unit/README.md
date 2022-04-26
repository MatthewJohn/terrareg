# Integration tests

These test against a semi-mocked models, testing against fake data supplied.

This is mainly to test 'user interfaces', i.e. the flask endpoints and flask-restful resources

Any functions that call the database directly (via sqlalchemy) are mocked.
