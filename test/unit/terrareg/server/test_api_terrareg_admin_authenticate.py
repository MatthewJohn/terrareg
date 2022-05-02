
import unittest.mock
import datetime

from test.unit.terrareg import (
    client, TerraregUnitTest
)



class TestApiTerraregAdminAuthenticate(TerraregUnitTest):

    def test_authenticated(self, client):
        """Test endpoint when user is authenticated."""
        cookie_expiry_mins = 5
        with unittest.mock.patch('terrareg.server.check_admin_authentication') as mock_admin_authentication:
            with unittest.mock.patch('terrareg.config.Config.SECRET_KEY', 'averysecretkey'):
                with unittest.mock.patch('terrareg.config.Config.ADMIN_SESSION_EXPIRY_MINS', cookie_expiry_mins):
                    # Update real app secret key
                    self.SERVER._app.secret_key = 'averysecretkey'

                    mock_admin_authentication.return_value = True

                    res = client.post('/v1/terrareg/auth/admin/login')
                    expected_cookie_expiry = datetime.datetime.now() - datetime.timedelta(minutes=cookie_expiry_mins)

                    assert res.status_code == 200
                    assert res.json == {'authenticated': True}
                    with client.session_transaction() as session:
                        # Assert that the session expiry is within 2 seconds of the expected expiry
                        assert (
                            (expected_cookie_expiry.timestamp() - session['expires'].timestamp()) < 2
                        )
                        assert session['is_admin_authenticated'] == True
                        assert len(session['csrf_token']) == 40

    def test_authenticated_without_secret_key(self, client):
        """Test endpoint and ensure session is not provided"""
        with unittest.mock.patch('terrareg.server.check_admin_authentication') as mock_admin_authentication:
            with unittest.mock.patch('terrareg.config.Config.SECRET_KEY', ''):
                # Update real app secret key with fake value,
                # otherwise an error would be received when checking the session.
                self.SERVER._app.secret_key = 'test'

                mock_admin_authentication.return_value = True

                res = client.post('/v1/terrareg/auth/admin/login')

                assert res.status_code == 403
                assert res.json == {'message': 'Sessions not enabled in configuration'}
                with client.session_transaction() as session:
                    # Assert that no session cookies were provided
                    assert 'expires' not in session
                    assert 'is_admin_authenticated' not in session
                    assert 'csrf_token' not in session

                # Update server secret to empty value and ensure a 403 is still received.
                # The session cannot be checked
                self.SERVER._app.secret_key = ''
                res = client.post('/v1/terrareg/auth/admin/login')

                assert res.status_code == 403
                assert res.json == {'message': 'Sessions not enabled in configuration'}

    def test_unauthenticated(self, client):
        """Test endpoint when user is authenticated."""
        with unittest.mock.patch('terrareg.server.check_admin_authentication') as mock_admin_authentication:

            mock_admin_authentication.return_value = False

            res = client.post('/v1/terrareg/auth/admin/login')

            assert res.status_code == 401
